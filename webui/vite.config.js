import {fileURLToPath, URL} from 'node:url'

import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig(({command, mode, ssrBuild}) => {
	const ret = {
		plugins: [vue()],
		resolve: {
			alias: {
				'@': fileURLToPath(new URL('./src', import.meta.url))
			}
		},
		// --- AGGIUNGI QUESTO BLOCCO SERVER ---
		server: {
			proxy: {
				// Reindirizza le richieste che iniziano con /avatars al backend
				'/avatars': {
					target: 'http://localhost:3000',
					changeOrigin: true,
				},
			}
		},
	};
	ret.define = {
		// Do not modify this constant, it is used in the evaluation.
		"__API_URL__": JSON.stringify("http://localhost:3000"),
	};
	return ret;
})