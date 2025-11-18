import react from '@vitejs/plugin-react-swc'
import UnoCSS from 'unocss/vite'

export default {
  plugins: [
    UnoCSS(),
    react(),
  ],
}