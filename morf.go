package main

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

const Suavizado_mon float64 = 0.25
const Suavizado_lex float64 = 0.75

type info_f struct {
	forma_en_el_texto string
	forma             string // pasada a minúscula si eso
	n_cat             int    // categorías posibles 0 => UNKN
	cat               string
	lem               string
	ras               string
	resuelta          bool
	cats              []string // posibles cats
}

var re_empieza_por_mayuscula = regexp.MustCompile(`^[A-ZÁÉÍÓÚÑ]`)
var re_empieza_por_minuscula = regexp.MustCompile(`^[a-záéíóúñ]`)
var re_cantidad = regexp.MustCompile(`\d+`)
var re_novalid = regexp.MustCompile(`[^\p{Latin}[[:digit:]][[:blank:]][[:punct:]][[:graph:]]«»—‹›“”"‘’‛]`)

func AnalizaTexto(entrada string, num_análisis string) string {
	if len(entrada) == 0 {
		return ""
	}
	//slog.Debug(num_análisis)
	entrada = re_novalid.ReplaceAllString(entrada, " ")
	frases := Segmenta_en_frases(entrada)
	salida := ""
	for _, f := range frases {
		salida += AnalizaFrase(f, num_análisis) + "\n"
	}

	return salida
}

func AnalizaFrase(entrada string, num_análisis string) string {

	slog.Debug("Analizando: [" + entrada + "]" + " " + num_análisis)

	if len(entrada) == 0 {
		return ""
	}

	// Tokenización
	frase_pretokenizada := strings.Replace(entrada, `\`, `\\`, -1)
	frase_tokenizada := TokenizaFrase(frase_pretokenizada) // minúscula principio de frase + multiwords

	slog.Debug("Tokenizacion: [" + frase_tokenizada + "]")

	lista_formas_texto := strings.Split(frase_tokenizada, " ")
	numero_de_formas := len(lista_formas_texto)

	info_formas := make([]info_f, numero_de_formas)

	for i := range info_formas {

		f := lista_formas_texto[i]
		info_formas[i].forma_en_el_texto = strings.Replace(f, "_", " ", -1)
	}

	Pon_forma_adecuada(info_formas) // Primera palabra a minuscula si eso

	Instancia(info_formas)

	Desambigua(info_formas)

	salida := Serializa_info_formas(info_formas)
	if num_análisis == "todos" {
		salida = Añade_posibles(salida)
	}
	return salida
}

func Añade_posibles(entrada string) string {
	salida := ""
	entrada = strings.TrimRight(entrada, "\n")
	ana := strings.Split(entrada, "\n")
	for _, a := range ana {
		partes := strings.Split(a, "/")
		forma := partes[0]
		//lema  := partes[1]
		resto := partes[2]

		di := Dicc[forma]
		n := len(di)
		añadidos := ""
		if n == 1 { // no ambigua
			salida += a + "\n"
		} else {
			cat := ""
			if strings.Contains(resto, ",") {
				prts := strings.Split(resto, ",")
				cat = prts[0]
				// fmt.Printf("Sal:[%s] forma:[%s] \t lema:[%s] \t resto:[%s]\n", a, forma, lema, resto)

			} else {
				cat = resto
			}
			//fmt.Printf("Sal:[%s] forma:[%s] \t lema:[%s] \t cat:[%s] \t resto:[%s]\n", a, forma, lema, cat, resto)
			for key, val := range di {
				// categorías distintas de la elegida en primer lugar
				if key != cat {
					// fmt.Println(key, val.lem, val.ras)
					restillo := ""
					if val.ras != "" {
						restillo += "," + val.ras
					}
					añadidos += fmt.Sprintf("\t%s/%s/%s%s", forma, val.lem, key, restillo)
				}
			}

			salida += a + añadidos + "\n"
		}
	}

	return salida
}

// Primera palabra a minúscula si eso
func Pon_forma_adecuada(info_formas []info_f) {

	primera_palabra_pasada := false
	for i := range info_formas {

		f := info_formas[i].forma_en_el_texto
		if primera_palabra_pasada {

			info_formas[i].forma = f
		} else {

			if re_empieza_por_minuscula.MatchString(f) || strings.Contains(f, `_`) {

				primera_palabra_pasada = true
				info_formas[i].forma = f
			} else if re_empieza_por_mayuscula.MatchString(f) {

				primera_palabra_pasada = true
				lower_f := strings.ToLower(f)
				consu_dic := ConsultaDiccionario((lower_f))
				if consu_dic == "" {
					info_formas[i].forma = f // Posible_nombre_propio
				} else {
					info_formas[i].forma = lower_f // a minúsculas primera palabra
				}
			} else {
				info_formas[i].forma = f
			}
		}
	}
}

func Serializa_info_formas(lista []info_f) string {

	salida := ""
	for _, l := range lista {
		lema := l.lem
		if len(lema) == 0 {
			lema = "UNKN"
		}
		if len(l.ras) > 0 {
			salida += fmt.Sprintf("%s/%s/%s,%s", l.forma_en_el_texto, lema, l.cat, l.ras)
		} else {
			//if len(l.forma_en_el_texto) > 0 {
			salida += fmt.Sprintf("%s/%s/%s", l.forma_en_el_texto, lema, l.cat)
			//}
		}

		salida += "\n"
	}
	return salida
}

func Instancia(info_formas []info_f) {

	re_sufijos_adj := regexp.MustCompile(`(.+)mente$`)
	//re_prefijos_n := regexp.MustCompile(`^euro(.+)`)

	for i := range info_formas {

		forma := info_formas[i].forma

		// Derivación
		derivada := false
		cate_deri := ""

		if !derivada {
			resu := re_sufijos_adj.FindStringSubmatch(forma)
			if len(resu) > 0 {
				form := resu[1]
				con := ConsultaDiccionario(form)
				res := strings.Index(con, "ADJ ")
				if res > -1 {
					derivada = true
					cate_deri = "ADV"
				}
			}
		}

		if derivada {
			info_formas[i].n_cat = 1
			info_formas[i].cat = cate_deri
			info_formas[i].lem = strings.ToUpper(forma)
			info_formas[i].ras = ""
			info_formas[i].resuelta = true
			continue
		}

		consulta := Dicc[forma]
		n_cats := len(consulta)

		// Posible NPR al principio de la frase
		añadir_npr := false
		forma_texto := info_formas[i].forma_en_el_texto
		if i == 0 && len(forma) > 3 && re_empieza_por_mayuscula.MatchString(forma_texto) {
			añadir_npr = true
		}

		lista_cats := make([]string, n_cats)

		if _, ok := consulta["AUX|V"]; ok { // son dos categorías

			lista_cats = make([]string, n_cats+1)
		}

		j := 0
		for k := range consulta {
			if k == "AUX|V" {

				lista_cats[j] = "AUX"
				j++
				lista_cats[j] = "V"
				j++

			} else {

				lista_cats[j] = k
				j++
			}
		}

		n_cats = len(lista_cats)
		// fmt.Println(consulta, n_cats, lista_cats)

		switch n_cats {
		case 0:

			if re_cantidad.MatchString(forma) {

				info_formas[i].n_cat = 1
				info_formas[i].cat = "Q"
				info_formas[i].lem = strings.ToUpper(forma)
				info_formas[i].ras = ""
				info_formas[i].resuelta = true

				// NPR
			} else if re_empieza_por_mayuscula.MatchString(forma) {

				info_formas[i].n_cat = 1
				info_formas[i].cat = "NPR"
				info_formas[i].lem = strings.ToUpper(forma)
				info_formas[i].ras = ""
				info_formas[i].resuelta = true

				// UNKN
			} else {

				info_formas[i].n_cat = 3
				info_formas[i].cat = "UNKN"
				info_formas[i].lem = strings.ToUpper(forma)
				info_formas[i].ras = ""
				info_formas[i].resuelta = false
				info_formas[i].cats = make([]string, 3)
				info_formas[i].cats[0] = "N"
				info_formas[i].cats[1] = "V"
				info_formas[i].cats[2] = "ADJ"
			}

		case 1:

			cat := lista_cats[0]

			if añadir_npr { // una mas npr

				info_formas[i].n_cat = n_cats + 1
				info_formas[i].cat = ""
				info_formas[i].lem = ""
				info_formas[i].ras = ""
				info_formas[i].resuelta = false
				info_formas[i].cats = make([]string, n_cats+1)
				copy(info_formas[i].cats[:], append(lista_cats, "NPR"))
				// fmt.Println(info_formas[i].cats)

			} else { // una sola
				info_formas[i].n_cat = 1
				info_formas[i].cat = cat
				info_formas[i].lem = Lem_de(forma, cat)
				info_formas[i].ras = Ras_de(forma, cat)
				info_formas[i].resuelta = true
			}

		default:

			info_formas[i].n_cat = n_cats
			info_formas[i].cat = ""
			info_formas[i].lem = ""
			info_formas[i].ras = ""
			info_formas[i].resuelta = false
			info_formas[i].cats = make([]string, n_cats)
			copy(info_formas[i].cats[:], lista_cats)
		}

	} // para cada forma
}

func Desambigua(info_formas []info_f) {

	para_resolver := ""
	for _, v := range info_formas {

		if v.resuelta {
			para_resolver += "1"
		} else {
			para_resolver += "0"
		}
	}

	re_ceros := regexp.MustCompile("0+")

	// hay ceros cuando hay que resolver
	for re_ceros.FindStringIndex(para_resolver) != nil {

		slog.Debug("Ambiguedad: " + para_resolver)

		loc := re_ceros.FindStringIndex(para_resolver)
		n_ceros := loc[1] - loc[0]

		Resuelve(info_formas, loc[0], loc[1])

		para_resolver = strings.Replace(para_resolver, "0", "1", n_ceros)

	}
}

func Resuelve(info_formas []info_f, a int, b int) {

	slog.Debug(fmt.Sprintf("%v", info_formas))

	número_posibilidades := 1
	i := a
	for i < b {
		número_posibilidades *= info_formas[i].n_cat
		i++
	}
	slog.Debug(fmt.Sprintf("%d %d -> posibilidades %d", a, b, número_posibilidades))

	if número_posibilidades > 1024 {
		Resuelve_por_secuencia(info_formas, a, b)
	} else {
		Calculo_exhaustivo(info_formas, a, b)
	}
}

func Calculo_exhaustivo(info_formas []info_f, a int, b int) {

	sequencias_posibles := Todas_las_sequencias(info_formas, a, b)
	cat_ant := "."
	cat_pos := "."
	if a > 0 {
		cat_ant = info_formas[a-1].cat
	}
	if b < len(info_formas) {
		cat_pos = info_formas[b].cat
	}
	prob_máxima := -1000000.
	sec_máxima := make([]string, b-a)
	sec_formas := make([]string, b-a)
	j := 0
	for i := a; i < b; i++ {
		sec_formas[j] = info_formas[j].forma
		j++
	}

	for _, seq := range sequencias_posibles {
		sequencia := strings.Split(seq, "-")
		prob := Probabilidad_de_esta(sequencia, cat_ant, cat_pos, sec_formas)
		if prob > prob_máxima {
			prob_máxima = prob
			sec_máxima = sequencia
		}
	}
	slog.Debug(fmt.Sprintf("\n\t\t\t\t\tElegida %v \t %4.2f\n", sec_máxima, prob_máxima))

	j = 0
	for i := a; i < b; i++ {
		cate := sec_máxima[j]
		forma := info_formas[i].forma
		info_formas[i].cat = cate
		info_formas[i].resuelta = true
		info_formas[i].lem = Lem_de(forma, cate)
		info_formas[i].ras = Ras_de(forma, cate)
		j++
	}
}

func Probabilidad_de_esta(sequencia []string, cat_ant string, cat_pos string, sec_formas []string) float64 {

	prob := 0.
	for k, cat := range sequencia {
		prob += Probabilidad_bigrama_de(cat_ant, cat, sec_formas[k])
		cat_ant = cat
	}
	prob += Big[cat_ant][cat_pos]
	slog.Debug(fmt.Sprintf("\t%v\t%4.2f", sequencia, prob))
	slog.Debug(fmt.Sprintf("\t\t\t\t %s %s \t %4.2f", cat_ant, cat_pos, Big[cat_ant][cat_pos]))
	return prob
}

func Probabilidad_bigrama_de(cat_ant string, cat string, forma string) float64 {

	mon := Mon[cat] * Suavizado_mon // monograma
	big := Big[cat_ant][cat]        // bigrama
	lex := Lex[forma][cat] * Suavizado_lex

	prob := mon + big + lex
	slog.Debug(fmt.Sprintf(" Pr(%s,%s) + Pr(%s) + Pr(%s, %s) %4.2f + %4.2f + %4.2f = %4.2f", cat_ant, cat, cat, cat, forma, big, mon, lex, prob))
	return prob
}

func Todas_las_sequencias(info_formas []info_f, a int, b int) []string {

	// Todas las secuencias posibles
	listas := make([][]string, b-a)

	i := 0
	lista_len_ant := 1

	for j := a; j < b; j++ { // bucle formas

		lista := info_formas[j].cats
		lista_len := len(lista)

		listas[i] = make([]string, lista_len_ant*lista_len) // formas posibles hasta esta
		ii := 0
		for k, v := range lista {
			if i == 0 {
				listas[i][k] = v
			} else {
				lista_ant := listas[i-1]
				for _, ll := range lista_ant {
					listas[i][ii] = ll + "-" + v
					ii++
				}
			}
		}
		//fmt.Println(i, "-", listas[i])
		i++
		lista_len_ant = lista_len * lista_len_ant
	}
	// Todas las posibilidades
	i--

	//fmt.Println(listas[i])
	return listas[i]
}

// no se calcula la probabilidad máxima
func Resuelve_por_secuencia(info_formas []info_f, a int, b int) {

	i := a
	for i < b {
		num_cat := info_formas[i].n_cat
		cat_ant := "."
		if i > 0 {
			cat_ant = info_formas[i-1].cat
		}

		j := 0
		prob_máxima := -100000.
		cat_máxima := ""
		forma := info_formas[i].forma
		for j < num_cat {

			categoría := info_formas[i].cats[j]
			probabilidad := Probabilidad_bigrama_de(cat_ant, categoría, forma)
			slog.Debug(fmt.Sprintf("\t\tP_mon(%s,%s|%s) = %5.2f", cat_ant, categoría, forma, probabilidad))
			if probabilidad > prob_máxima {
				prob_máxima = probabilidad
				cat_máxima = categoría
			}
			j++
		}
		slog.Debug(fmt.Sprintf("\n\t\tElegida [%s/%s]", forma, cat_máxima))

		info_formas[i].cat = cat_máxima
		info_formas[i].resuelta = true
		if info_formas[i].lem == "" {
			info_formas[i].lem = Lem_de(forma, cat_máxima)
			info_formas[i].ras = Ras_de(forma, cat_máxima)
		}
		i++
	}
}
