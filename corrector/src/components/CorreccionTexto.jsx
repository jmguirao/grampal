import { useState, useEffect } from 'react';

import axios from "axios"

const FONDO_SI    = "papayawhip"
const FONDO_NO    = "peachpuff"


export default function CorreccionTexto({paraCorregir}) {

  const [data, setData] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

	const cambiarFondo = (evt) => {
		evt.target.style.backgroundColor = FONDO_SI
		// const col = evt.target.dataset.col
		// const row = evt.target.dataset.row
		const parent  = evt.target.parentElement
		const sibilings = Array.from(parent.children).filter( s => s !== evt.target)
		sibilings.forEach( s => s.style.backgroundColor = FONDO_NO)
	}

	const nada = () => {}


	const colsDe = (f, max, fila) => {
		let filacol = []
		const colu = f.split("\t")
		// se rellena hasta el máximo de columnas con blancos
		const colus = Array.from({length: max}, (_, index) => colu[index] || '')
		let i=0
		for (const c of colus) {
			let fondo 
			fondo = i == 0 ? FONDO_SI : FONDO_NO
			const clickFunc = c.includes("/") ? cambiarFondo : nada
			const d = c.replaceAll("/", " /	 ").replaceAll(",", ", ")
			filacol.push(<td key={i} className='px-4 py-2' style={{backgroundColor:fondo}} data-col={i} data-row={fila}
			                         onClick={clickFunc}>{d}</td>)
			i++
		}
		return filacol
	}

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await axios.put(import.meta.env.VITE_URL_GP,  
				            { texto: paraCorregir},
										{ headers: 
											{'Content-Type': 'application/x-www-form-urlencoded; charset=utf-8'}
										});
        setData(response.data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  if (loading) return <div>Cargando análisis...</div>;
  if (error) return <div>{error}</div>
	const filas = data.trim().split("\n")
	//console.log(filas)
	let cols = 0
	// columnas
	for (const i of filas) {
		let x = i.split("\t").length
		if (x > cols) { cols = x}
	}
	//console.log(cols)
	// todo bien
	return (
		<div className="py-2 px-5  h-auto d-flex flex-column" style={{backgroundColor:'#ffd6ba'}}>
			 <div className='lead px-4 py-3' style={{fontSize: '80%'}}>Click en la opción correcta</div>
			<table><tbody  className='' style={{fontSize:'75%'}} id="tbody">
				{
					filas.map((f, i) => {
						return (<tr key={i}>
						          {[... colsDe(f, cols, i)]}
						        </tr>)
					})
				}
			</tbody></table>
			<br/>
			<button class="mt-3 bg-white hover:bg-gray-100 px-1 border border-gray-400 rounded shadow"
			         onClick={Corregido}>
         Corregido
      </button>			
		</div>
	)
}

const Corregido = () => {
	const corregidos = []
	console.log("Corregidos: ")
	const tbody = document.getElementById('tbody')
	for (const tr of tbody.children) {
		for (const td of tr.children) {
			if (td.style.backgroundColor === FONDO_SI) {
				corregidos.push(td.innerText.replaceAll(" / ", "/"))
			}
		}
	}
	let resu = ""
	for (const c of corregidos) {
		resu += c + "<br>"
	}
	const newTab = window.open('', '_blank')
	if (newTab) {
		newTab.document.write(`
		<!DOCTYPE html>
		<html>
			<head>
				<title>Texto corregido</title>
				<style>
					body { 
						font-family: Arial, sans-serif; 
						padding: 20px;
						font-size: 120%;
						line-height: 1.3;
						background-color: #f8f8f8;
					}
				</style>
			</head>
			<body contenteditable="true">
			${resu}
			</body>
		</html>		
		`)
		newTab.document.close()
	} 
}

// import useSWR from "swr"

// const fetcher = (url) => axios.get(url).then(res => res.data)

	// const { data, error, isLoading } = useSWR(`http://localhost:8001/${paraCorregir}`, fetcher)
	
	// if (isLoading) return (
	// 	<div>Cargando ...</div>
	// )
	
	// if (error) return (
	// 	<div>`Error: ${error}`</div>
	// )

	// if (data) {
	// console.log(data)
// }