import { existsSync } from 'node:fs';
import { resolve } from 'node:path';
import { build, type Plugin } from 'vite';
import vue from '@vitejs/plugin-vue';

export function benchmarkReportPlugin(root: string): Plugin | undefined {
    const entry = resolve(root, 'src/enterprise/views/ai-benchmark/report/main.ts');
    if (!existsSync(entry)) return;
    const id = 'virtual:benchmark-report';
    const resolvedID = `\0${id}`;
    const dependencies = new Set<string>();
    return {
        name: 'benchmark-report',
        resolveId(source) {
            if (source === id) return resolvedID;
        },
        async load(source) {
            if (source !== resolvedID) return;
            const result = await build({
                root,
                configFile: false,
                publicDir: false,
                logLevel: 'warn',
                resolve: { preserveSymlinks: true },
                plugins: [
                    vue(),
                    {
                        name: 'benchmark-report-dependencies',
                        buildEnd() {
                            for (const file of this.getModuleIds()) {
                                if (!file.includes('node_modules') && !file.startsWith('\0')) {
                                    dependencies.add(file.split('?')[0]);
                                }
                            }
                        },
                    },
                ],
                define: {
                    'process.env.NODE_ENV': JSON.stringify('production'),
                    __VUE_OPTIONS_API__: true,
                    __VUE_PROD_DEVTOOLS__: false,
                    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: false,
                },
                build: {
                    write: false,
                    minify: 'oxc',
                    cssCodeSplit: false,
                    lib: { entry, name: 'BenchmarkReport', formats: ['iife'] },
                },
            });
            const outputs = (Array.isArray(result) ? result : [result]).flatMap((bundle) =>
                'output' in bundle ? bundle.output : [],
            );
            const script = outputs
                .filter((output) => output.type === 'chunk')
                .map((output) => output.code)
                .join('\n');
            const styles = outputs
                .filter((output) => output.type === 'asset' && output.fileName.endsWith('.css'))
                .map((output) => (output.type === 'asset' ? String(output.source) : ''))
                .join('\n');
            if (!script || !styles) throw new Error('Benchmark report bundle is incomplete');
            for (const file of dependencies) this.addWatchFile(file);
            return `export const script = ${JSON.stringify(script)}; export const styles = ${JSON.stringify(styles)};`;
        },
        handleHotUpdate(context) {
            if (!dependencies.has(context.file)) return;
            const module = context.server.moduleGraph.getModuleById(resolvedID);
            if (module) context.server.moduleGraph.invalidateModule(module);
            context.server.ws.send({ type: 'full-reload' });
        },
    };
}
