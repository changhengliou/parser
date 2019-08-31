#include <cstdio>
#include <cstdlib>
#include <iostream>
#include <stack>
#include <string.h>
#include <string>
#include <unistd.h>

using namespace std;

/*
 * Convert infix regexp re to postfix notation.
 * Insert . as explicit concatenation operator.
 * Cheesy parser, return static buffer.
 */
char *re2post(char *re) {
  int nalt, natom;
  static char buf[8000];
  char *dst;
  struct {
    int nalt;
    int natom;
  } paren[100], *p;

  p = paren;
  dst = buf;
  nalt = 0;
  natom = 0;
  if (strlen(re) >= sizeof buf / 2)
    return nullptr;
  for (; *re; re++) {
    switch (*re) {
    case '(':
      if (natom > 1) {
        --natom;
        *dst++ = '.';
      }
      if (p >= paren + 100)
        return nullptr;
      p->nalt = nalt;
      p->natom = natom;
      p++;
      nalt = 0;
      natom = 0;
      break;
    case '|':
      if (natom == 0)
        return nullptr;
      while (--natom > 0)
        *dst++ = '.';
      nalt++;
      break;
    case ')':
      if (p == paren)
        return nullptr;
      if (natom == 0)
        return nullptr;
      while (--natom > 0)
        *dst++ = '.';
      for (; nalt > 0; nalt--)
        *dst++ = '|';
      --p;
      nalt = p->nalt;
      natom = p->natom;
      natom++;
      break;
    case '*':
    case '+':
    case '?':
      if (natom == 0)
        return nullptr;
      *dst++ = *re;
      break;
    default:
      if (natom > 1) {
        --natom;
        *dst++ = '.';
      }
      *dst++ = *re;
      natom++;
      break;
    }
  }
  if (p != paren)
    return nullptr;
  while (--natom > 0)
    *dst++ = '.';
  for (; nalt > 0; nalt--)
    *dst++ = '|';
  *dst = 0;
  return buf;
}

/*
 * Represents an NFA state plus zero or one or two arrows exiting.
 * if c == Match, no arrows out; matching state.
 * If c == Split, unlabeled arrows to out and out1 (if != nullptr).
 * If c < 256, labeled arrow with character c to out.
 */
enum { Match = 256, Split = 257 };
int nstate;
struct State {
  int c;
  State *next;
  State *next2;
  int lastlist;
  State(int c) : c(c) {}
  State(int c, State *next, State *next2) : c(c), next(next), next2(next2) {
    nstate++;
  }
};
/* matching state */
State matchstate(Match);

/*
 * Since the out pointers in the list are always
 * uninitialized, we use the pointers themselves
 * as storage for the Ptrlists.
 */
union Ptrlist {
  Ptrlist *next;
  State *state;
};

/*
 * A partially built NFA without the matching state filled in.
 * NfaFrag.start points at the start state.
 * NfaFrag.out is a list of places that need to be set to the
 * next state for this fragment.
 */
struct Frag {
  State *start;
  Ptrlist *next;
  Frag() {}
  Frag(State *start, Ptrlist *next) : start(start), next(next) {}
};

/* Create singleton list containing just outp. */
Ptrlist *newList(State **outp) {
  Ptrlist *list;

  list = (Ptrlist *)outp;
  list->next = nullptr;
  return list;
}

/* Patch the list of states at out to point to start. */
void patch(Ptrlist *list, State *state) {
  Ptrlist *next;

  for (; list; list = next) {
    next = list->next;
    list->state = state;
  }
}

/* Join the two lists l1 and l2, returning the combination. */
Ptrlist *append(Ptrlist *list, Ptrlist *list2) {
  Ptrlist *head;

  head = list;
  while (list->next != nullptr) {
    list = list->next;
  }
  list->next = list2;
  return head;
}

/*
 * Convert postfix regular expression to NFA.
 * Return start state.
 */
State *post2nfa(string postfix) {
  stack<Frag> fragSatck;
  Frag e1, e2, e;
  State *state;

  if (postfix.empty())
    return nullptr;

  for (char p : postfix) {
    switch (p) {
    default:
      state = new State(p, nullptr, nullptr);
      fragSatck.push(Frag(state, newList(&state->next)));
      break;
    case '.': /* catenate */
      e2 = fragSatck.top();
      fragSatck.pop();
      e1 = fragSatck.top();
      fragSatck.pop();
      patch(e1.next, e2.start);
      fragSatck.push(Frag(e1.start, e2.next));
      break;
    case '|': /* alternate */
      e2 = fragSatck.top();
      fragSatck.pop();
      e1 = fragSatck.top();
      fragSatck.pop();
      state = new State(Split, e1.start, e2.start);
      fragSatck.push(Frag(state, append(e1.next, e2.next)));
      break;
    case '?': /* zero or one */
      e = fragSatck.top();
      fragSatck.pop();
      state = new State(Split, e.start, nullptr);
      fragSatck.push(Frag(state, append(e.next, newList(&state->next2))));
      break;
    case '*': /* zero or more */
      e = fragSatck.top();
      fragSatck.pop();
      state = new State(Split, e.start, nullptr);
      patch(e.next, state);
      fragSatck.push(Frag(state, newList(&state->next2)));
      break;
    case '+': /* one or more */
      e = fragSatck.top();
      fragSatck.pop();
      state = new State(Split, e.start, nullptr);
      patch(e.next, state);
      fragSatck.push(Frag(e.start, newList(&state->next2)));
      break;
    }
  }

  e = fragSatck.top();
  fragSatck.pop();
  if (!fragSatck.empty()) {
    return nullptr;
  }

  patch(e.next, &matchstate);
  return e.start;
}

struct List {
  State **s;
  int n;
};
List l1, l2;
static int listid;

void addstate(List *, State *);
void step(List *, int, List *);

/* Compute initial state list */
List *startlist(State *start, List *l) {
  l->n = 0;
  listid++;
  addstate(l, start);
  return l;
}

/* Check whether state list contains a match. */
int ismatch(List *l) {
  int i;

  for (i = 0; i < l->n; i++)
    if (l->s[i] == &matchstate)
      return 1;
  return 0;
}

/* Add s to l, following unlabeled arrows. */
void addstate(List *l, State *s) {
  if (s == nullptr || s->lastlist == listid)
    return;
  s->lastlist = listid;
  if (s->c == Split) {
    /* follow unlabeled arrows */
    addstate(l, s->next);
    addstate(l, s->next2);
    return;
  }
  l->s[l->n++] = s;
}

/*
 * Step the NFA from the states in clist
 * past the character c,
 * to create next NFA state set nlist.
 */
void step(List *clist, int c, List *nlist) {
  int i;
  State *s;

  listid++;
  nlist->n = 0;
  for (i = 0; i < clist->n; i++) {
    s = clist->s[i];
    if (s->c == c)
      addstate(nlist, s->next);
  }
}

/* Run NFA to determine whether it matches s. */
int match(State *start, char *s) {
  int c;
  List *clist, *nlist, *t;

  clist = startlist(start, &l1);
  nlist = &l2;
  for (; *s; s++) {
    c = *s & 0xFF;
    step(clist, c, nlist);
    t = clist;
    clist = nlist;
    nlist = t; /* swap clist, nlist */
  }
  return ismatch(clist);
}

int main(int argc, char **argv) {
  int i;
  char *post;
  State *start;
  const char *regex = "(ab)*c";

  post = re2post((char *)regex);
  start = post2nfa(post);
  if (start == nullptr) {
    fprintf(stderr, "error in post2nfa %s\n", post);
    return 1;
  }

  l1.s = (State **)malloc(nstate * sizeof l1.s[0]);
  l2.s = (State **)malloc(nstate * sizeof l2.s[0]);
  for (i = 2; i < argc; i++)
    if (match(start, argv[i]))
      printf("%s\n", argv[i]);
  return 0;
}
