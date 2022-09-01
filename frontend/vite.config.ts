import { defineConfig, loadEnv, ConfigEnv, UserConfig } from 'vite';
import { createHtmlPlugin } from 'vite-plugin-html';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import { wrapperEnv } from './src/utils/get-env';
import { visualizer } from 'rollup-plugin-visualizer';
import viteCompression from 'vite-plugin-compression';
import VueSetupExtend from 'vite-plugin-vue-setup-extend';
import eslintPlugin from 'vite-plugin-eslint';
import vueJsx from '@vitejs/plugin-vue-jsx';
import DefineOptions from 'unplugin-vue-define-options/vite';
import MonacoEditorPlugin from 'vite-plugin-monaco-editor';

// @see: https://vitejs.dev/config/
export default defineConfig(({ mode }: ConfigEnv): UserConfig => {
    const env = loadEnv(mode, process.cwd());
    const viteEnv = wrapperEnv(env);

    return {
        // base: "/",
        // alias config
        resolve: {
            alias: {
                '@': resolve(__dirname, './src'),
                'vue-i18n': 'vue-i18n/dist/vue-i18n.cjs.js',
            },
        },
        // global css
        css: {
            preprocessorOptions: {
                scss: {
                    additionalData: `@use "@/styles/var.scss" as *;`,
                },
            },
        },
        server: {
            port: viteEnv.VITE_PORT,
            open: viteEnv.VITE_OPEN,
            host: '0.0.0.0',
            // https: false,
            proxy: {
                '/api/v1': {
                    target: 'http://localhost:9999',
                    changeOrigin: true,
                    // rewrite: (path) => path.replace(/^\/api/, ''),
                },
            },
        },
        plugins: [
            vue(),
            createHtmlPlugin({
                inject: {
                    data: {
                        title: viteEnv.VITE_GLOB_APP_TITLE,
                    },
                },
            }),
            DefineOptions(),
            // * EsLint 报错信息显示在浏览器界面上
            eslintPlugin({
                exclude: ['**/*.js'],
            }),
            // * vite 可以使用 jsx/tsx 语法
            vueJsx(),
            // * name 可以写在 script 标签上
            VueSetupExtend(),

            MonacoEditorPlugin({}),
            viteEnv.VITE_REPORT && visualizer(),
            // * gzip compress
            viteEnv.VITE_BUILD_GZIP &&
                viteCompression({
                    verbose: true,
                    disable: false,
                    threshold: 10240,
                    algorithm: 'gzip',
                    ext: '.gz',
                }),
        ],
        esbuild: {
            pure: viteEnv.VITE_DROP_CONSOLE ? ['console.log', 'debugger'] : [],
        },
        build: {
            outDir: 'dist',
            minify: 'esbuild',
            rollupOptions: {
                output: {
                    // Static resource classification and packaging
                    chunkFileNames: 'assets/js/[name]-[hash].js',
                    entryFileNames: 'assets/js/[name]-[hash].js',
                    assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
                },
            },
        },
    };
});
