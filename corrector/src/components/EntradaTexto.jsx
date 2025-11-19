
export default function EntradaTexto({handleEnviar}) {

	return (
		<div id="contenido" class="py-4 px-14" className="px-5 pt-3 h-auto d-flex flex-column" style={{backgroundColor:'#ffd6ba', height:'100%'}}>

			<span class="" className="px-8 py-2">Texto para etiquetar: &nbsp; </span>
      
			<form class="bg-white text-xs mt-2" className="w-100">
				<textarea id="formu" rows="22" cols="120" class="px-1 py-1" 
				          className="w-100 px-2 py-1" style={{backgroundColor:'white'}}/>
			</form>
			<button 
			        className="btn btn-warning my-2"
			        onClick={() => {const formu = document.getElementById('formu'); handleEnviar(formu.value)}}>
         Enviar
      </button>
			
			<br/>	
		</div>

	)
}
