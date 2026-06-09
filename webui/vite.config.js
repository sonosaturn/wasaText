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
		server: {
			proxy: {
				'/avatars': { target: 'http://localhost:3000', changeOrigin: true },
				'/images':  { target: 'http://localhost:3000', changeOrigin: true },
			}
		},
		preview: {
			proxy: {
				'/avatars': { target: 'http://localhost:3000', changeOrigin: true },
				'/images':  { target: 'http://localhost:3000', changeOrigin: true },
			}
		},
	};
	ret.define = {
		// Do not modify this constant, it is used in the evaluation.
		"__API_URL__": JSON.stringify("http://localhost:3000"),
		"__VUE_PROD_HYDRATION_MISMATCH_DETAILS__": false,
	};
	return ret;
})