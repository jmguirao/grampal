package main

import (
	"regexp"
	"strings"
)

func Segmenta_en_frases(texto string) []string {

	re_seg_1 := regexp.MustCompile(`(\.) ([A-Z])`)
	re_seg_11 := regexp.MustCompile(`(\.) *\n+ *([A-Z])`)
	re_seg_2 := regexp.MustCompile(`(;) (.)`)
	re_seg_3 := regexp.MustCompile(`(:) (.)`)

	texto = re_seg_1.ReplaceAllString(texto, "$1||$2")
	texto = re_seg_11.ReplaceAllString(texto, "$1||$2")
	texto = re_seg_2.ReplaceAllString(texto, "$1||$2")
	texto = re_seg_3.ReplaceAllString(texto, "$1||$2")

	re_bb := regexp.MustCompile(`\|\||\n\n+`) // frase o párrafo

	salida := re_bb.Split(texto, -1)

	return salida
}

func TokenizaFrase(frase string) string {

	//	salida := ReconoceCantidades(frase)
	salida := SeparaPuntuacion(frase)      // por espacios
	salida = ReconoceNPRsMultiword(salida) // uniendo tokens_con_subs
	salida = ReconoceMultiwordsTrie(salida)
	salida = SeparaAmalgamas(salida)
	salida = SeparaClíticos(salida)
	return salida
}

// func ReconoceCantidades(entrada string) string { // en un futuro
// 	return entrada
// }

func compatible_clitico(form string) bool {

	con := Dicc[form]
	_, ok := con["AUX|V"]
	if ok {
		ras := Ras_de(form, "AUX|V")
		if strings.Contains(ras, "imper") {
			return true
		}
		if strings.Contains(ras, "inf") {
			return true
		}
	}

	_, ok = con["V"]
	if ok {
		ras := Ras_de(form, "V")
		if strings.Contains(ras, "imper") {
			return true
		}
		if strings.Contains(ras, "inf") {
			return true
		}
	}

	return false
}

func quita_tildes(pal string) string {
	sal := strings.Replace(pal, "á", "a", 1)
	sal = strings.Replace(sal, "é", "e", 1)
	sal = strings.Replace(sal, "í", "i", 1)
	sal = strings.Replace(sal, "ó", "o", 1)
	sal = strings.Replace(sal, "ú", "u", 1)
	return sal
}

func SeparaClíticos(entrada string) string {

	re_cli := regexp.MustCompile(`(.+)(le|me|se|te)(lo)?$`)

	salida := ""
	formas := strings.Split(entrada, " ")
	for _, v := range formas {

		clitico := false
		resu := re_cli.FindStringSubmatch(v)
		if len(resu) > 0 {
			form := resu[1]

				// quita tildes del principio
				if (len(resu) > 3) {
					form = quita_tildes(form)
			}

			if compatible_clitico(form) {
				clitico = true
			}
		}
		if clitico {
			salida += resu[1] + " " + resu[2] + " "
			if resu[3] != "" {
				salida += resu[3] + " "
			}
			salida = quita_tildes(salida)
		} else {
			salida += v + " "
		}
	}
	return strings.TrimSpace(salida)
}

// por espacios
func SeparaPuntuacion(entrada string) string {

	re_puntuacion := regexp.MustCompile(`[?¿¡!.,;:<>()\[\]{}"'«»“”]`)
	re_spsp := regexp.MustCompile(`\s+`)

	salida := re_puntuacion.ReplaceAllStringFunc(entrada, func(m string) string {
		return " " + m + " "
	})

	return strings.TrimSpace(re_spsp.ReplaceAllString(salida, " "))
}

// y pone un signo _ entre las palbras para formar un solo token
func ReconoceNPRsMultiword(entrada string) string {

	re_NPR := regexp.MustCompile(`([A-Z][a-záéíóúñü]+(?: (?:del? )?(?:la )?(?:[A-Z][a-záéíóúñü]+))*)`)
	re_spsp := regexp.MustCompile(`\s+`)

	salida := re_NPR.ReplaceAllStringFunc(entrada, func(m string) string {
		return " " + strings.Replace(m, " ", "_", -1) + " "
	})

	salida = strings.TrimSpace(re_spsp.ReplaceAllString(salida, " "))
	return salida
}

// al y del
func SeparaAmalgamas(frase string) string {

	re_al := regexp.MustCompile(`\bal\b`)
	re_del := regexp.MustCompile(`\bdel\b`)

	frase = re_al.ReplaceAllString(frase, "a el")
	frase = re_del.ReplaceAllString(frase, "de el")

	return frase
}

// * Devuelve separado con _ en multiwords
func ReconoceMultiwordsTrie(frase string) string {

	// primera letra minúscula
	frase_t := strings.ToLower(frase[:1]) + frase[1:]

	// no words , return
	i := strings.Index(frase_t, " ")
	if i < 0 {
		return frase
	}

	m, _, b := Mw.LongestPrefix(frase_t)

	// mw sola
	if (b && m == frase_t) {
		return strings.ReplaceAll(frase, " ", "_")
	}

	espacios := findSpaces(frase)
	
	// log.Tracef("\n\n                                frase_t:[%s]\n", frase_t)
	// log.Trace(espacios)

	espacios = append([]int{0}, espacios...)
	for i , posi := range espacios {
		//log.Tracef("                                     %d  %d  [%s]", i, posi, frase_t[posi:])
		// principio de la frase
		suplemento := 1
		if i == 0 { suplemento = 0}
		
		m, _, b := Mw.LongestPrefix(frase_t[posi+suplemento:])
		if b && strings.HasPrefix(frase_t[posi+suplemento:]+" ", m+" ") {
			para_substituir := strings.ReplaceAll(m, " ", "_")
			frase = strings.Replace(frase, m, para_substituir, 1)
		}
		// log.Tracef("                %t frase: [%s] m: [%s] [%s]", b, frase, m+" ", frase_t[posi+suplemento:]+" ")
			
		// me salto la última palabra
		if i >= len(espacios)-2 {break}
	}

	return frase
}

func findSpaces(s string) []int {
	var positions []int
	for i, char := range s {
		if char == ' ' {
			positions = append(positions, i)
		}
	}
	return positions
}
