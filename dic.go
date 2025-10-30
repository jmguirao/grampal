package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/armon/go-radix"
)

const DATADIR string = "data/"
const DICCIONARIO string = "formas_es"

type dict_ent struct {
	lem, ras string
}

var Dicc = make(map[string]map[string]dict_ent) // Diccionario de formas, categorías, lemas, rasgos
var Mw = radix.New()

var Mon = make(map[string]float64)            // Probabilidades de monogramas
var Big = make(map[string]map[string]float64) // Probabilidades de bigramas
var Lex = make(map[string]map[string]float64) // Probabilidades lexico|cat

func Lee_Diccionario_desde_texto(data_dir, diccionario string) {
	dic_file := path.Join(DATADIR, DICCIONARIO)
	diccTextFile, err := os.Open(dic_file + ".txt")
	if err != nil {
		slog.Error("cargando diccionario: " + err.Error())
		os.Exit(1)
	}
	defer diccTextFile.Close()

	var dicc_t = make(map[string]string) // Diccionario en texto
	input := bufio.NewScanner(diccTextFile)
	for input.Scan() { // Se lee como texto asociado a la forma
		l := strings.Split(input.Text(), "/")
		if len(l) > 2 {
			_, ok := dicc_t[l[0]]
			if ok {
				dicc_t[l[0]] += "\n" + strings.Join(l[1:], "/") // ambiguas
			} else {
				dicc_t[l[0]] = strings.Join(l[1:], "/") // primera entrada
			}
		}
	}
	// Paso a extructura de datos
	for forma, info := range dicc_t {
		Dicc[forma] = make(map[string]dict_ent)
		variantes := strings.Split(info, "\n")

		// Multiwords
		n_espacios := strings.Count(forma, " ")
		if n_espacios > 0 {
			Mw.Insert(forma, "")
		}
		for _, v := range variantes {
			var e dict_ent
			var cat string
			lineas := strings.Split(v, "/") //  forma/LEMA/rasgos en el diccionario

			if len(lineas) > 0 {
				e.lem = lineas[0]
			}
			if len(lineas) > 1 {
				cat = lineas[1]
			}
			if len(lineas) > 2 {
				e.ras = lineas[2]
			}
			Dicc[forma][cat] = e
		}
	}
	// Signos de puntuación
	signos_de_puntuacion := `?¿¡!.,;:<>()[]{}"'«»“”/%`
	for _, c := range signos_de_puntuacion {
		signo := string(c)
		Dicc[signo] = make(map[string]dict_ent)
		Dicc[signo]["PUNCT"] = dict_ent{lem: signo}
	}
}

func CargaDatos() error {
	Lee_Diccionario_desde_texto(DATADIR, DICCIONARIO)
	Lee_Monogramas()
	Lee_Bigramas()
	Lee_Modelo_lexico()	
	fmt.Println("Datos cargados")
	return nil
	// return errors.New("errorr aqui")
}

func ConsultaDiccionario(entrada string) string {
	consulta := Dicc[entrada]

	salida := ""
	for k, v := range consulta {

		salida += k + " "
		if v.ras != "" {
			salida += v.ras + " "
		}
		salida += v.lem + "\n"
	}
	if salida == "" {
		return "UNKN " + entrada
	}
	return salida
}

func Ras_de(palabra string, cat string) string {
	consulta := Dicc[palabra]

	if cat == "AUX" {
		cat = "AUX|V"
	}

	if cat == "V" && consulta[cat].ras == "" {
		cat = "AUX|V"
	}

	return consulta[cat].ras
}

func Lem_de(palabra string, cat string) string {

	if cat == "NPR" {
		return strings.ToUpper(palabra)
	}

	consulta := Dicc[palabra]
	if cat == "AUX" {
		cat = "AUX|V"
	}

	if cat == "V" && consulta[cat].lem == "" {
		cat = "AUX|V"
	}

	return consulta[cat].lem
}

func Lee_Monogramas() {

	//	dic_file := path.Join(os.Getenv("GOPATH")+DATADIR, MODELO)
	dic_file := path.Join(DATADIR, "General")
	ModMonFile, err := os.Open(dic_file + ".mon")
	if err != nil {
		slog.Error(fmt.Sprintf("Problemas con el archivo de monogramas: %v", err))
		os.Exit(1)
	}
	defer ModMonFile.Close()
	input := bufio.NewScanner(ModMonFile)

	re_split := regexp.MustCompile(`\s+`)
	re_noletras := regexp.MustCompile(`[@%;,:!#]`)

	total_cats := 0
	total_tok := 0

	for input.Scan() {
		l := input.Text()
		if !re_noletras.MatchString(l) {

			arr := re_split.Split(l, -1)
			if len(arr) == 2 {

				cat := arr[0]
				cue, _ := strconv.Atoi(arr[1])

				if len(cat) > 0 && cue > 0 {

					total_cats++
					total_tok += cue
					Mon[cat] += float64(cue)
				}
			}
		}
	} // end for

	slog.Info(fmt.Sprintf("Categorias %d, tokens %d", total_cats, total_tok))

	log := math.Log10(float64(total_tok))
	for k, v := range Mon {

		Mon[k] = math.Log10(v) - log
		//fmt.Println(k, v, Mon[k])
	}
}

func Lee_Bigramas() {

	//	dic_file := path.Join(os.Getenv("GOPATH")+DATADIR, MODELO)
	dic_file := path.Join(DATADIR, "General")
	ModBigFile, err := os.Open(dic_file + ".big")
	if err != nil {
		slog.Error(fmt.Sprintf("Problemas con el archivo de bigramas: %v", err))
	}

	defer ModBigFile.Close()
	input := bufio.NewScanner(ModBigFile)

	re_split := regexp.MustCompile(`\s+`)
	re_split_c := regexp.MustCompile(`-`)
	re_noletras := regexp.MustCompile(`[@%;,:!#]`)

	total_cats := 0
	total_tok := 0

	var cuentas = make(map[string]float64)

	for input.Scan() {

		l := input.Text()
		if !re_noletras.MatchString(l) {

			arr := re_split.Split(l, -1)
			if len(arr) == 2 {

				cats := arr[0]
				cue, _ := strconv.Atoi(arr[1])

				if len(cats) > 0 && cue > 0 {

					arr = re_split_c.Split(cats, -1)
					cat1 := arr[0]
					cat2 := arr[1]

					total_cats++
					total_tok += cue

					if Big[cat1] == nil {
						Big[cat1] = make(map[string]float64)
					}

					Big[cat1][cat2] += float64(cue)
					cuentas[cat2] += float64(cue)
				}
			}
		}
	} // end for

	slog.Info(fmt.Sprintf("Categorias bigrama %d, tokens %d", total_cats, total_tok))

	for cat2, v := range cuentas {

		log := math.Log10(float64(v))

		for cat1 := range Big {
			if Big[cat1][cat2] < 1. {
				Big[cat1][cat2] = 0.5
			}
			Big[cat1][cat2] = math.Log10(Big[cat1][cat2]) - log
		}
	}
}

func Lee_Modelo_lexico() {

	re_noletras := regexp.MustCompile(`[@%;,.!#]`)
	re_split := regexp.MustCompile(`[/:]`)

	var cuentas = make(map[string]int)

	//	dic_file := path.Join(os.Getenv("GOPATH")+DATADIR, MODELO)
	dic_file := path.Join(DATADIR, "General")
	ModLexFile, err := os.Open(dic_file + ".lex")
	if err != nil {
		slog.Error(fmt.Sprintf("Problemas con el modelo estadístico léxico: %v", err))
	}
	defer ModLexFile.Close()

	input := bufio.NewScanner(ModLexFile)
	for input.Scan() {
		l := input.Text()
		if !re_noletras.MatchString(l) {
			arr := re_split.Split(l, -1)
			form := arr[0]
			cat := arr[1]
			cue, _ := strconv.Atoi(arr[2])
			cuentas[cat] += cue

			consulta := Dicc[form]
			n_cats := len(consulta)
			if n_cats > 1 {
				//fmt.Println(form, cat, cue, consulta)
				if Lex[form] == nil {
					Lex[form] = make(map[string]float64)
				}
				Lex[form][cat] = float64(cue)
			}
		}
	}
	//log := math.Log10(float64(total_tok))
	for f := range Lex {
		for c := range Lex[f] {
			Lex[f][c] = math.Log10(Lex[f][c]) - math.Log10(float64(cuentas[c]))
			//fmt.Println(f, c, Lex[f][c])

		}
		consulta := Dicc[f]
		for k := range consulta {
			if Lex[f][k] == 0. && k != "AUX|V" {
				Lex[f][k] = math.Log10(0.5) - math.Log10(float64(cuentas[k]))
				//fmt.Println("\t", k, Lex[f][k])
			}
		}
	}
}
