package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/parser"
)

const Prompt = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, Prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		lex := lexer.New(line)
		p := parser.New(lex)
		prog := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		fmt.Fprintf(out, "%s\n", prog.String())
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
