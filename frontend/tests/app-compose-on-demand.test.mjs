import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { baseParse } from '@vue/compiler-dom';
import { parse } from '@vue/compiler-sfc';
import ts from 'typescript';
import { computed, proxyRefs, reactive, ref } from 'vue';

const view = (name) =>
    parse(readFileSync(new URL(`../src/views/app-store/installed/${name}/index.vue`, import.meta.url), 'utf8'))
        .descriptor;
const upgrade = view('upgrade');
const target = (version = '2.0', detailId = 2, dockerCompose = 'services:\n  app:\n    image: demo:2') => ({
    version,
    detailId,
    dockerCompose,
});

function drawer(overrides = {}) {
    const calls = { params: [], versions: [], errors: [], operations: [], confirmations: [] };
    const bindings = {
        ref,
        reactive,
        computed,
        onBeforeUnmount: () => {},
        defineEmits: () => () => {},
        defineExpose: () => {},
        useGlobalStore: () => ({ currentNode: ref('local') }),
        Rules: { requiredSelect: {} },
        i18n: { global: { t: (key) => key } },
        bus: { on() {}, off() {}, emit() {} },
        isOpenclawCurrentHTTPVersion: () => false,
        isOpenclawHTTPSWindowVersion: () => false,
        uuidv4: () => 'task-id',
        MsgError: (message) => calls.errors.push(message),
        MsgSuccess() {},
        ElMessageBox: {
            confirm: (...args) => {
                calls.confirmations.push(args);
                return Promise.resolve();
            },
        },
        getAppStoreConfig: async () => ({ data: { upgradeBackup: 'Disable', upgradeDeleteImage: 'Disable' } }),
        getAppInstallParams: async (...args) => {
            calls.params.push(args);
            return { data: { isEdit: true, dockerCompose: 'original custom instance compose' } };
        },
        getAppUpdateVersions: async (...args) => {
            calls.versions.push(args);
            return { data: [target()] };
        },
        installedOp: async (...args) => calls.operations.push(args),
        ignoreUpgrade: async () => {},
        ...overrides,
    };
    const ast = ts.createSourceFile('upgrade.ts', upgrade.scriptSetup.content, ts.ScriptTarget.Latest, true);
    const script = ast.statements
        .filter((statement) => !ts.isImportDeclaration(statement))
        .map((statement) => statement.getText(ast))
        .join('\n');
    const compiled = ts.transpileModule(script, { compilerOptions: { target: ts.ScriptTarget.ESNext } }).outputText;
    const state = new Function(
        ...Object.keys(bindings),
        `${compiled}\nreturn { acceptParams, getVersions, handleClose, openDiff, onOperate, operate, versions, loading, loadingVersions, composeReady, currentComposeReady, operateReq, oldContent, newContent, newCompose, useNewCompose, isEdit, node, appInstallID, feishuLegacy, feishuConfigured, open };`,
    )(...Object.values(bindings));
    state.node.value = 'child & node';
    state.appInstallID.value = 17;
    state.oldContent.value = 'saved instance compose';
    return { ...state, calls };
}

function findNode(node, predicate) {
    if (predicate(node)) return node;
    for (const child of node.children || []) {
        const found = findNode(child, predicate);
        if (found) return found;
    }
}

const template = baseParse(upgrade.template.content);
const button = findNode(template, (node) =>
    node.props?.some((prop) => prop.name === 'on' && prop.exp?.content === 'onOperate'),
);
const disabled = button.props.find((prop) => prop.name === 'bind' && prop.arg.content === 'disabled').exp.content;
const isDisabled = (state) => new Function('state', `with (state) { return (${disabled}); }`)(proxyRefs(state));

test('upgrade validates the current template and loads only the default target on the selected node', async () => {
    const state = drawer();
    assert.equal(isDisabled(state), true);
    await state.getVersions('');
    assert.deepEqual(state.calls.params, [[17, 'child & node']]);
    assert.deepEqual(state.calls.versions, [[{ appInstallID: 17, loadCompose: true }, 'child & node']]);
    assert.equal(state.isEdit.value, true, 'authoritative configuration overrides the stale list flag');
    assert.equal(state.oldContent.value, 'original custom instance compose');
    assert.equal(state.operateReq.detailId, 2);
    assert.equal(state.newContent.value, target().dockerCompose);
    assert.equal(state.composeReady.value, true);
    assert.equal(isDisabled(state), false);
});

test('older agents without isEdit retain the existing custom Compose warning', async () => {
    const state = drawer({ getAppInstallParams: async () => ({ data: { dockerCompose: 'old agent compose' } }) });
    state.isEdit.value = true;
    await state.getVersions('');
    assert.equal(state.isEdit.value, true);
    assert.equal(state.oldContent.value, 'old agent compose');
    assert.equal(state.composeReady.value, true);
});

test('current-template failure blocks preview and submission without changing the saved Compose', async () => {
    const state = drawer({
        getAppInstallParams: async () => {
            throw new Error('template unavailable');
        },
    });
    await state.getVersions('');
    assert.equal(state.calls.versions.length, 0);
    assert.equal(state.oldContent.value, 'saved instance compose');
    assert.equal(state.currentComposeReady.value, false);
    assert.equal(state.loadingVersions.value, false);
    assert.equal(isDisabled(state), true);
    await state.onOperate();
    await state.operate();
    assert.equal(state.calls.confirmations.length, 0);
    assert.equal(state.calls.operations.length, 0);
});

test('a failed target switch cannot submit the previous target and can be retried', async () => {
    let fail = true;
    const state = drawer({
        getAppUpdateVersions: async (req) => {
            if (req.updateVersion === '3.0' && fail) throw new Error('download failed');
            return { data: [target(), target('3.0', 3, 'services:\n  app:\n    image: demo:3')] };
        },
    });
    await state.getVersions('');
    state.operateReq.version = '3.0';
    await state.getVersions('3.0');
    assert.equal(state.operateReq.detailId, 2);
    assert.equal(state.oldContent.value, 'original custom instance compose');
    assert.equal(isDisabled(state), true);
    await state.onOperate();
    assert.equal(state.calls.confirmations.length, 0);
    fail = false;
    await state.getVersions('3.0');
    assert.equal(state.operateReq.detailId, 3);
    assert.equal(state.composeReady.value, true);
    assert.equal(state.calls.params.length, 1, 'switching targets reuses the verified current configuration');
    assert.equal(isDisabled(state), false);
});

test('empty and missing target Compose cannot become valid previews', async () => {
    for (const data of [[target('2.0', 2, '')], [target('2.0', 2, '  ')], [target('different', 3)]]) {
        const state = drawer({ getAppUpdateVersions: async () => ({ data }) });
        await state.getVersions('2.0');
        assert.equal(state.composeReady.value, false);
        assert.equal(state.operateReq.detailId, 0);
        assert.equal(state.oldContent.value, 'original custom instance compose');
        assert.deepEqual(state.calls.errors, ['container.contentEmpty']);
    }
});

test('out-of-order responses cannot overwrite the latest selection or enable it too soon', async () => {
    const pending = new Map();
    const state = drawer({
        getAppUpdateVersions: (req) => new Promise((resolve) => pending.set(req.updateVersion, resolve)),
    });
    state.currentComposeReady.value = true;
    const first = state.getVersions('2.0');
    const latest = state.getVersions('3.0');
    pending.get('2.0')({ data: [target()] });
    await first;
    assert.equal(state.loadingVersions.value, true);
    assert.equal(state.composeReady.value, false);
    pending.get('3.0')({ data: [target('3.0', 3)] });
    await latest;
    assert.equal(state.operateReq.version, '3.0');
    assert.equal(state.operateReq.detailId, 3);
    assert.equal(state.composeReady.value, true);
});

test('closing a drawer invalidates in-flight template responses', async () => {
    let resolve;
    const state = drawer({ getAppUpdateVersions: () => new Promise((done) => (resolve = done)) });
    state.currentComposeReady.value = true;
    const request = state.getVersions('2.0');
    state.handleClose();
    resolve({ data: [target()] });
    await request;
    assert.equal(state.open.value, false);
    assert.equal(state.composeReady.value, false);
    assert.equal(state.operateReq.detailId, 0);
});

test('a late settings response cannot reopen a closed upgrade drawer', async () => {
    let resolve;
    const state = drawer({ getAppStoreConfig: () => new Promise((done) => (resolve = done)) });
    const request = state.acceptParams({ id: 17 }, 'upgrade');
    state.handleClose();
    resolve({ data: {} });
    await request;
    assert.equal(state.open.value, false);
    assert.equal(state.calls.params.length, 0);
    assert.equal(state.calls.versions.length, 0);
});

test('ignore version requests skip both current and target Compose loading', async () => {
    const state = drawer();
    state.operateReq.operate = 'ignore';
    await state.getVersions('');
    assert.equal(state.calls.params.length, 0);
    assert.equal(state.calls.versions[0][0].loadCompose, false);

    const source = view('ignore/create').scriptSetup.content;
    const ast = ts.createSourceFile('ignore.ts', source, ts.ScriptTarget.Latest, true);
    const declaration = ast.statements
        .filter(ts.isVariableStatement)
        .flatMap((statement) => statement.declarationList.declarations)
        .find((item) => item.name.getText(ast) === 'getVersions');
    const compiled = ts.transpileModule(`const getVersions = ${declaration.initializer.getText(ast)};`, {
        compilerOptions: { target: ts.ScriptTarget.ESNext },
    }).outputText;
    let request;
    const getVersions = new Function(
        'getAppUpdateVersions',
        'appInstallID',
        'versions',
        'req',
        `${compiled}\nreturn getVersions;`,
    )(
        async (req) => {
            request = req;
            return { data: [target()] };
        },
        ref(17),
        ref([]),
        {},
    );
    await getVersions();
    assert.deepEqual(request, { appInstallID: 17, loadCompose: false });
});

test('blank custom Compose is rejected before the upgrade confirmation', async () => {
    const state = drawer();
    await state.getVersions('');
    state.useNewCompose.value = true;
    state.newCompose.value = '  ';
    await state.onOperate();
    assert.deepEqual(state.calls.errors, ['container.contentEmpty']);
    assert.equal(state.calls.confirmations.length, 0);
});

test('Feishu legacy transition still enforces the required backup', async () => {
    const state = drawer({
        getAppUpdateVersions: async () => ({ data: [{ ...target(), feishuLegacy: true, feishuConfigured: true }] }),
    });
    state.operateReq.backup = false;
    await state.getVersions('');
    assert.equal(state.feishuLegacy.value, true);
    assert.equal(state.feishuConfigured.value, true);
    assert.equal(state.operateReq.backup, true);
});

test('explicit current-configuration loading preserves child-node routing', () => {
    const source = readFileSync(new URL('../src/api/modules/app.ts', import.meta.url), 'utf8');
    const ast = ts.createSourceFile('app-api.ts', source, ts.ScriptTarget.Latest, true);
    const declaration = ast.statements
        .filter(ts.isVariableStatement)
        .flatMap((statement) => statement.declarationList.declarations)
        .find((item) => item.name.getText(ast) === 'getAppInstallParams');
    const compiled = ts.transpileModule(`const getAppInstallParams = ${declaration.initializer.getText(ast)};`, {
        compilerOptions: { target: ts.ScriptTarget.ESNext },
    }).outputText;
    const request = new Function('http', `${compiled}\nreturn getAppInstallParams;`)({ get: (url) => url });
    assert.equal(request(17), 'apps/installed/params/17');
    assert.equal(request(17, 'child & node'), 'apps/installed/params/17?operateNode=child%20%26%20node');
});
