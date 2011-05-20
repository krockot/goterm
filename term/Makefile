include $(GOROOT)/src/Make.inc
TARG=term
GOFILES:=termdefs.go exectty.go
CGOFILES:=term.go
CGO_OFILES:=goterm.o

include $(GOROOT)/src/Make.pkg

DOLLAR:="$"

termdefs.go: termdefs.c
	godefs `echo -n $(CGO_FLAGS) | sed 's/\(^ ^$(DOLLAR)]*\)/-f \1/g'` -g term termdefs.c > termdefs.go
	gofmt -w termdefs.go


