Query Engine
============

This part contains more sophisticated logic.

```
command.go -      commands are final result of parsing.
language.peg -    query language grammar definition.
language.peg.go - go file generated from language.peg
node.go -         syntax tree nodes used during the parser.
parser.go -       support code used by the parser. Parser entry point
predicate.go -    logic for predicate node.
```
