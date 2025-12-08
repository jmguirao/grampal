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

func SeparaClíticos(entrada string) string {

	re_cli := regexp.MustCompile(`(.+)(le|me|se|te)(lo)?$`)

	salida := ""
	formas := strings.Split(entrada, " ")
	for _, v := range formas {

		clitico := false
		resu := re_cli.FindStringSubmatch(v)
		if len(resu) > 0 {
			form := resu[1]
			if compatible_clitico(form) {
				clitico = true
			}
		}
		if clitico {
			salida += resu[1] + " " + resu[2] + " "
			if resu[3] != "" {
				salida += resu[3] + " "
			}
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

func ReconoceMultiwordsTrie(frase string) string {

	frase_t := strings.ToLower(frase[:1]) + frase[1:]

	i := strings.Index(frase_t, " ")
	if i < 0 {
		return frase
	}

	// devuelve frase con subrallados si es mw
	last_trie := ""
	for i > 0 {

		m, _, b := Mw.LongestPrefix(frase_t)
		last_trie = m
		log.Debug("m: ", m, "  b: ", b)
		if b {

			r := strings.Replace(m, " ", "_", -1)

			if strings.Contains(frase, m) { // corregido bug
				frase = strings.Replace(frase, m, r, 1)
			} else { // primera mayúscula
				mm := strings.ToUpper(m[:1]) + m[1:]
				rr := strings.Replace(mm, " ", "_", -1)
				frase = strings.Replace(frase, mm, rr, 1)
			}

			frase_t = strings.Replace(frase_t, m, r, 1)
			frase_t = frase_t[i+1:]
		}
		i = strings.Index(frase_t, " ")
		frase_t = frase_t[i+1:]
	}
	log.Debug("frase: ", frase)
	// para corrregir conflicto: a la pared -- a la par
	if (len(frase) != len(last_trie)) {frase = strings.Replace(frase, "_", " ", -1)}
	return frase
}
