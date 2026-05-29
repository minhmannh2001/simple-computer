package asm

import "fmt"

type marker interface{ placeholder() }

type LABEL struct{ Name string }

func (LABEL) placeholder()          {}
func (l LABEL) String() string      { return l.Name }

type SYMBOL struct{ Name string }

func (SYMBOL) placeholder()         {}
func (s SYMBOL) String() string     { return "%" + s.Name }

type NUMBER struct{ Value uint16 }

func (NUMBER) placeholder()         {}
func (n NUMBER) String() string     { return fmt.Sprintf("0x%X", n.Value) }
