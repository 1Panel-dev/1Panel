import { defineConfig, loadEnv, ConfigEnv, UserConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import { wrapperEnv } from './src/utils/get-env';
import { visualizer } from 'rollup-plugin-visualizer';
import viteCompression from 'vite-plugin-compression';
import VueSetupExtend from 'vite-plugin-vue-setup-extend';
import eslintPlugin from 'vite-plugin-eslint';
import vueJsx from '@vitejs/plugin-vue-jsx';
import DefineOptions from 'unplugin-vue-define-options/vite';

import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';
import svgLoader from 'vite-svg-loader';

const prefix = `monaco-editor/esm/vs`;

export default defineConfig(({ mode }: ConfigEnv): UserConfig => {
    const env = loadEnv(mode, process.cwd());
    const viteEnv = wrapperEnv(env);

    return {
        resolve: {
            alias: {
                '@': resolve(__dirname, './src'),
                'vue-i18n': 'vue-i18n/dist/vue-i18n.cjs.js',
                xpack: resolve(__dirname, './src/xpack'),
            },
        },
        css: {
            preprocessorOptions: {
                scss: {
                    additionalData: `@use "@/styles/var.scss" as *;`,
                    api: 'modern',
                },
            },
        },
        server: {
            port: viteEnv.VITE_PORT,
            open: viteEnv.VITE_OPEN,
            host: '0.0.0.0',
            proxy: {
                '/api/v1': {
                    target: 'http://localhost:9999/',
                    changeOrigin: true,
                    ws: true,
                },
            },
        },
        plugins: [
            vue(),
            DefineOptions(),
            eslintPlugin({
                exclude: ['**/*.js'],
            }),
            vueJsx(),
            VueSetupExtend(),
            viteEnv.VITE_REPORT && visualizer(),
            viteEnv.VITE_BUILD_GZIP &&
                viteCompression({
                    verbose: true,
                    disable: false,
                    threshold: 10240,
                    algorithm: 'gzip',
                    ext: '.gz',
                }),
            AutoImport({
                resolvers: [
                    ElementPlusResolver({
                        importStyle: 'sass',
                    }),
                ],
            }),
            Components({
                resolvers: [
                    ElementPlusResolver({
                        importStyle: 'sass',
                    }),
                ],
            }),
            svgLoader({
                defaultImport: 'url',
            }),
        ],
        esbuild: {
            pure: viteEnv.VITE_DROP_CONSOLE ? ['console.log'] : [],
            drop: viteEnv.VITE_DROP_CONSOLE && process.env.NODE_ENV === 'production' ? ['debugger'] : [],
        },
        build: {
            outDir: '../cmd/server/web',
            minify: 'esbuild',
            target: 'esnext',
            rollupOptions: {
                output: {
                    chunkFileNames: 'assets/js/[name]-[hash].js',
                    entryFileNames: 'assets/js/[name]-[hash].js',
                    assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
                    manualChunks: {
                        jsonWorker: [`${prefix}/language/json/json.worker`],
                        cssWorker: [`${prefix}/language/css/css.worker`],
                        htmlWorker: [`${prefix}/language/html/html.worker`],
                        tsWorker: [`${prefix}/language/typescript/ts.worker`],
                        editorWorker: [`${prefix}/editor/editor.worker`],
                    },
                },
            },
        },
    };
});
