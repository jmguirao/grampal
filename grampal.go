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
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

const MAX_FRAS_LENGTH int = 4096
const MAX_TEXT_LENGTH int = 4194304

func Servicio_diccionario(w http.ResponseWriter, r *http.Request) {

	fras := r.URL.Path[1:]
	if len(fras) > 32 {
		io.WriteString(w, "")
	} else {
		io.WriteString(w, ConsultaDiccionario(fras))
	}
}

func Servicio_etiquetador(w http.ResponseWriter, r *http.Request) {

	// Como etiquetador
	if r.Method == "POST" {
		r.ParseForm()
		texto := r.FormValue("texto")
		if len(texto) > MAX_TEXT_LENGTH {
			io.WriteString(w, fmt.Sprintf("Excedido longitud máxima de texto: %d", MAX_TEXT_LENGTH))
		} else {
			io.WriteString(w, AnalizaTexto(texto, "uno"))
		}
	}

	// etiquedador todas las opciones (para corrector)
	if r.Method == "PUT" {
		r.ParseForm()
		texto := r.FormValue("texto")
		if len(texto) > MAX_TEXT_LENGTH {
			io.WriteString(w, fmt.Sprintf("Excedido longitud máxima de texto: %d", MAX_TEXT_LENGTH))
		} else {
			io.WriteString(w, AnalizaTexto(texto, "todos"))
		}
	}

	// GET, etiquedador para frases
	fras := r.URL.Path[1:]
	if len(fras) > MAX_FRAS_LENGTH {
		io.WriteString(w, fmt.Sprintf("Excedido longitud máxima de texto: %d", MAX_FRAS_LENGTH))
	} else {
		if len(fras) > 0 {
			io.WriteString(w, AnalizaTexto(fras, num_análisis))
		}
	}
}

var num_análisis string = "uno"
var log *logrus.Logger

func main() {

	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true, // Show full timestamp instead of elapsed time
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false, // Enable colors for better readability
		ForceColors:     false, // Force colors even when not in a terminal
	})

	dictPtr := flag.Bool("dic", false, "Uso como diccionario")
	serPtr := flag.Bool("ser", false, "Servicio")
	portPtr := flag.String("port", "8001", "Puerto")
	todosPtr := flag.Bool("todos", false, "Todos los análisis")
	corsPtr := flag.Bool("cors", false, "Para desarrollo (cors)")
	debugPtr := flag.Bool("debug", false, "Modo Debug")
	tracePtr := flag.Bool("trace", false, "Modo Trace")

	flag.Parse()

	if *debugPtr {
		log.SetLevel(logrus.DebugLevel)
	}
	if *tracePtr {
		log.SetLevel(logrus.TraceLevel)
	}
  

	log.WithField("Log Level", log.GetLevel()).Info()


	// log.WithField("user_id", 123).Info("hello from logrus")
	// a := "dsfsdf"
	// log.Debugf("%s hello from logrus", a)
	

	funciona_como := "etiquedador"
	if (*dictPtr) {
		funciona_como = "diccionario"
	}

	log.WithField("Funciona como", funciona_como).Info()

	err := CargaDatos(funciona_como)
	if err != nil {
		log.WithError(err).Fatal("Cargando datos")
	}

	if *todosPtr {
		num_análisis = "todos"
	}

	if *serPtr { // Funciona como servicio

		puerto := *portPtr
		mux := http.NewServeMux()

		if *dictPtr {
			log.Infof("Servicio como diccionario en el puerto %s", puerto)
			mux.HandleFunc("/", Servicio_diccionario)
		} else {
			log.Infof("Servicio como etiquedador en el puerto %s", puerto)
			mux.HandleFunc("/", Servicio_etiquetador)
		}

		// cors middleware
		c := cors.New(cors.Options{})
		if *corsPtr {
			c = cors.New(cors.Options{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut},
				AllowCredentials: true,
			})
		}
		handler := c.Handler(mux)
		http.ListenAndServe(":"+puerto, handler)

	} else if funciona_como == "diccionario" {
		Bucle_entrada_teclado("Palabra", "uno")
	} else if funciona_como == "etiquedador" {
		Bucle_entrada_teclado("Frase", num_análisis)
	}
}

func Bucle_entrada_teclado(prompt string, num_análisis string) {

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
				log.Debugf("Análisis de  [%s]\n\n", entrada)

				salida := ""
				switch prompt {
				case "Frase":
					salida = AnalizaTexto(entrada, num_análisis)
				case "Palabra":
					salida = ConsultaDiccionario(entrada)
				}
				fmt.Printf("%s\n\n", salida)
			}
		}
	}
}
