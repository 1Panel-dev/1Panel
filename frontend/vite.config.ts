import { defineConfig, loadEnv, ConfigEnv, UserConfig } from 'vite';
import { createHtmlPlugin } from 'vite-plugin-html';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import { wrapperEnv } from './src/utils/getEnv';
import { visualizer } from 'rollup-plugin-visualizer';
import viteCompression from 'vite-plugin-compression';
import VueSetupExtend from 'vite-plugin-vue-setup-extend';
import eslintPlugin from 'vite-plugin-eslint';
import vueJsx from '@vitejs/plugin-vue-jsx';
import importToCDN from 'vite-plugin-cdn-import';
// import AutoImport from "unplugin-auto-import/vite";
// import Components from "unplugin-vue-components/vite";
// import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
import DefineOptions from 'unplugin-vue-define-options/vite';

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
        // server config
        server: {
            host: '0.0.0.0', // 服务器主机名，如果允许外部访问，可设置为"0.0.0.0"
            port: viteEnv.VITE_PORT,
            open: viteEnv.VITE_OPEN,
            cors: true,
            // https: false,
            // 代理跨域（mock 不需要配置，这里只是个事列）
            proxy: {
                '/api': {
                    // target: "https://www.fastmock.site/mock/f81e8333c1a9276214bcdbc170d9e0a0", // fastmock
                    target: 'https://mock.mengxuegu.com/mock/629d727e6163854a32e8307e', // easymock
                    changeOrigin: true,
                    rewrite: (path) => path.replace(/^\/api/, ''),
                },
            },
        },
        // plugins
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
            eslintPlugin(),
            // * vite 可以使用 jsx/tsx 语法
            vueJsx(),
            // * name 可以写在 script 标签上
            VueSetupExtend(),
            // * demand import element(如果使用了cdn引入,没必要使用element自动导入了)
            // AutoImport({
            // 	resolvers: [ElementPlusResolver()]
            // }),
            // Components({
            // 	resolvers: [ElementPlusResolver()]
            // }),
            // * cdn 引入（vue、element-plus）
            importToCDN({
                modules: [
                    // vue按需引入会导致依赖vue的插件出现问题(列如:pinia/vuex)
                    // {
                    // 	name: "vue",
                    // 	var: "Vue",
                    // 	path: "https://unpkg.com/vue@next"
                    // },
                    // 使用cdn引入element-plus时,开发环境还是需要在main.js中引入element-plus,可以不用引入css
                    // {
                    // 	name: "element-plus",
                    // 	var: "ElementPlus",
                    // 	path: "https://unpkg.com/element-plus",
                    // 	css: "https://unpkg.com/element-plus/dist/index.css"
                    // }
                ],
            }),
            // * 是否生成包预览
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
        // build configure
        build: {
            outDir: 'dist',
            // esbuild 打包更快，但是不能去除 console.log
            minify: 'esbuild',
            // minify: "terser",
            // terserOptions: {
            // 	compress: {
            // 		drop_console: viteEnv.VITE_DROP_CONSOLE,
            // 		drop_debugger: true
            // 	}
            // },
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
