
import EntradaTexto from "./components/EntradaTexto"
import CorreccionTexto from "./components/CorreccionTexto"

import { useState } from "react"

function App() {

	const [texto, setTexto] = useState('')
	const [corrección, setCorrección] = useState(true)


	const handleEnviar = (texto) => {
		setTexto(texto)
		setCorrección(false)
	} 

	
  return (
    <div style={{backgroundColor:'#ffd6ba'}} class="h100">
			<div class="w-full top-0 left-0 px-8 py-3" style={{backgroundColor:'#ffc6aa'}}>
				<img className="w-25" src="/public/Uam_LLI.png" alt="logo" />
				<span class="text-lg font-bold font-stretch-[260%] font-sans" className="fw-bold ps-5"> Corrector </span>
			</div>
				{ corrección ? (
					<EntradaTexto handleEnviar={handleEnviar}/>
				) : (
					<CorreccionTexto paraCorregir={texto}/>
				)
        }
    </div>
  )
}

export default App
