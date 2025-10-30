/*
|
| GRAMPAL
| =======
|
| jmguirao@ugr.es nov-16
| re-born         oct-25
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

func main() {

	dictPtr := flag.Bool("dic", false, "Uso como diccionario")
	flag.Parse()

	funciona_como := "etiquedador"
	if *dictPtr {
		funciona_como = "diccionario"
	}
	fmt.Printf("Funcionando como: %s\n", funciona_como)

	err := CargaDatos()
	if err != nil {
		slog.Error("cargando datos: " + err.Error())
		os.Exit(1)
	}

	if funciona_como == "diccionario" {
		Bucle_entrada_teclado("Palabra")
	}
	if funciona_como == "etiquedador" {
		Bucle_entrada_teclado("Frase")
	}

}

func Bucle_entrada_teclado(prompt string) {

	re_spsp := regexp.MustCompile(`\s+`)
	var entrada string
	re_permitidos := regexp.MustCompile(`^[A-ZÁÉÍÓÚÜÑa-záéíóúñ <>()\[\]{}"'«»“”.,;:0-9—–\-?¿!¡%]+$`)

	for {
		fmt.Printf(prompt + ": ")
		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			fmt.Println(err)
		} else {
			entrada = strings.TrimSpace(string(line))
			if len(entrada) > 2048 {
				fmt.Println("Demasiado larga")
			} else if entrada == "" {
				
			} else if !re_permitidos.MatchString(entrada) {
				fmt.Println("Caractéres no permitidos")
			} else {
				entrada = re_spsp.ReplaceAllString(entrada, " ")
				entrada = strings.Trim(entrada, " ")
				fmt.Printf("Análisis de  [%s]\n\n", entrada)

				salida := ""
				switch prompt {
				case "Frase":
					salida = AnalizaTexto(entrada)
				case "Palabra":
					salida = ConsultaDiccionario(entrada)
				}
				fmt.Printf("%s\n\n", salida)
			}
		}
	}
}
