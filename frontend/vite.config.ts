import { resolve } from 'path';
import { wrapperEnv } from './src/utils/get-env';
import { visualizer } from 'rollup-plugin-visualizer';
import viteCompression from 'vite-plugin-compression';
import eslintPlugin from 'vite-plugin-eslint2';
import vueJsx from '@vitejs/plugin-vue-jsx';
import { defineConfig, loadEnv, ConfigEnv, UserConfig, Plugin } from 'vite';
import vue from '@vitejs/plugin-vue';
import pkg from './package.json';
import dayjs from 'dayjs';

import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';
import svgLoader from 'vite-svg-loader';

function patchCodeFilterOverflow(): Plugin {
    return {
        name: 'patch-code-filter-overflow',
        configResolved(config) {
            const target = config.plugins.find((p) => p.name === 'vite:asset-import-meta-url');
            if (target?.transform && typeof target.transform === 'object' && target.transform.filter) {
                (target.transform.filter as Record<string, any>).code = 'import.meta.url';
            }
        },
    };
}

const prefix = `monaco-editor/esm/vs`;

function getManualChunkName(id: string): string | undefined {
    if (id.includes(`${prefix}/language/json/json.worker`)) return 'jsonWorker';
    if (id.includes(`${prefix}/language/css/css.worker`)) return 'cssWorker';
    if (id.includes(`${prefix}/language/html/html.worker`)) return 'htmlWorker';
    if (id.includes(`${prefix}/language/typescript/ts.worker`)) return 'tsWorker';
    if (id.includes(`${prefix}/editor/editor.worker`)) return 'editorWorker';

    if (id.includes('node_modules/echarts/')) return 'vendor-echarts';
    if (id.includes('node_modules/element-plus/') || id.includes('node_modules/@element-plus/')) {
        return 'vendor-element-plus';
    }
    if (id.includes('node_modules/codemirror/') || id.includes('node_modules/@codemirror/')) {
        return 'vendor-codemirror';
    }

    return undefined;
}

const { dependencies, devDependencies, name, version } = pkg;
const __APP_INFO__ = {
    pkg: { dependencies, devDependencies, name, version },
    lastBuildTime: dayjs().format('YYYY-MM-DD HH:mm:ss'),
};

export default defineConfig(({ mode }: ConfigEnv): UserConfig => {
    const env = loadEnv(mode, process.cwd());
    const viteEnv = wrapperEnv(env);
    const isProduction = mode === 'production';

    return {
        resolve: {
            preserveSymlinks: true,
            alias: {
                '@': resolve(__dirname, './src'),
                'vue-i18n': 'vue-i18n/dist/vue-i18n.cjs.js',
            },
        },
        define: {
            __APP_INFO__: JSON.stringify(__APP_INFO__),
        },
        css: {
            preprocessorOptions: {
                scss: {
                    additionalData: `@use "@/styles/var.scss" as *;`,
                    silenceDeprecations: ['legacy-js-api'],
                },
            },
        },
        server: {
            port: viteEnv.VITE_PORT,
            open: viteEnv.VITE_OPEN,
            host: '0.0.0.0',
            sourcemapIgnoreList: (sourcePath) => {
                return sourcePath.includes('node_modules');
            },
            proxy: {
                '/api/v2': {
                    target: 'http://localhost:9999/',
                    changeOrigin: true,
                    ws: true,
                },
            },
        },
        plugins: [
            patchCodeFilterOverflow(),
            vue(),
            eslintPlugin({
                cache: false,
                exclude: ['**/*.js'],
            }),
            vueJsx(),
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
                imports: ['vue', 'vue-router'],
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
        build: {
            sourcemap: false,
            outDir: '../core/cmd/server/web',
            minify: 'oxc',
            target: 'esnext',
            cssCodeSplit: false,
            rollupOptions: {
                output: {
                    chunkFileNames: 'assets/js/[name]-[hash].js',
                    entryFileNames: 'assets/js/[name]-[hash].js',
                    assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
                    minify: {
                        compress: {
                            dropDebugger: viteEnv.VITE_DROP_CONSOLE && isProduction,
                            treeshake: {
                                manualPureFunctions: viteEnv.VITE_DROP_CONSOLE ? ['console.log'] : [],
                            },
                        },
                        mangle: true,
                    },
                    manualChunks: getManualChunkName,
                },
            },
        },
    };
});
