import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { baseParse, compile } from '@vue/compiler-dom';
import { parse } from '@vue/compiler-sfc';
import { ElAvatar } from 'element-plus/es/components/avatar/index.mjs';
import ts from 'typescript';
import * as Vue from 'vue';
import { isRemoteAppIcon, resolveAppIconUrl } from '../src/utils/app-icon.ts';

const cdnURL = 'https://cdn.example/icons/php.png?v=abc123&token=a%2Bb#icon';
const apiURL = '/api/v2';
const getAppIconUrl = (key, node, icon = '') => resolveAppIconUrl(key, node, icon, apiURL);

test('accepts absolute HTTP(S) image URLs without changing their query or fragment', () => {
    for (const icon of [
        cdnURL,
        'http://cdn.example/icon.png',
        'HTTPS://cdn.example:443/icon.png',
        'https://[::1]/icon',
    ]) {
        assert.equal(isRemoteAppIcon(icon), true, icon);
        assert.equal(getAppIconUrl('php', 'child-node', icon), icon);
    }
});

test('rejects unsafe schemes, credentials, normalization ambiguities and malformed URLs', () => {
    for (const icon of [
        '//cdn.example/icon.png',
        'https:cdn.example/icon.png',
        'https:///icon.png',
        'javascript:alert(1)',
        'data:image/png;base64,eA==',
        'file:///tmp/icon.png',
        'https://user:password@cdn.example/icon.png',
        'https://@cdn.example/icon.png',
        'https://cdn.example:bad/icon.png',
        'https://cdn.example:65536/icon.png',
        'https://cdn.example/%zz',
        'https://cdn.example/icon.png?token=%zz',
        ' https://cdn.example/icon.png',
        'https://cdn.example/icon.png\n',
        'https://cdn.example/\u00a0icon.png',
        'https://cdn.example/\u0085icon.png',
        'https://cdn.example\\other/icon.png',
    ]) {
        assert.equal(isRemoteAppIcon(icon), false, icon);
        assert.equal(getAppIconUrl('php', 'child-node', icon), '/api/v2/apps/icon/php?operateNode=child-node');
    }
});

test('keeps historical icons and missing fields on the original node-aware API', () => {
    for (const icon of ['', undefined, 'app_php.png?etag=%22old%22', 'aGlzdG9yaWNhbC1pbWFnZQ==']) {
        assert.equal(
            getAppIconUrl('local-php', 'child & node', icon),
            '/api/v2/apps/icon/local-php?operateNode=child%20%26%20node',
        );
    }
    assert.equal(getAppIconUrl('php'), '/api/v2/apps/icon/php');
});

const readView = (name) => parse(readFileSync(new URL(`../src/views/${name}`, import.meta.url), 'utf8')).descriptor;

function launcherResolver() {
    const source = readView('home/app/index.vue').scriptSetup.content;
    const ast = ts.createSourceFile('launcher.ts', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
    const declaration = ast.statements
        .filter(ts.isVariableStatement)
        .flatMap((statement) => statement.declarationList.declarations)
        .find((item) => item.name.getText(ast) === 'getAppIconSrc');
    const expression = ts.transpileModule(`const getAppIconSrc = ${declaration.initializer.getText(ast)};`, {
        compilerOptions: { target: ts.ScriptTarget.ESNext },
    }).outputText;
    return new Function('isRemoteAppIcon', 'getAppIconUrl', `${expression}\nreturn getAppIconSrc;`)(
        isRemoteAppIcon,
        getAppIconUrl,
    );
}

test('home launcher uses CDN URLs and retains inline historical Base64 icons', () => {
    const getAppIconSrc = launcherResolver();
    assert.equal(getAppIconSrc({ key: 'php', icon: cdnURL }), cdnURL);
    assert.equal(
        getAppIconSrc({ key: 'php', icon: 'aGlzdG9yaWNhbC1pbWFnZQ==' }),
        'data:image/png;base64,aGlzdG9yaWNhbC1pbWFnZQ==',
    );
    assert.equal(getAppIconSrc({ key: 'php', icon: 'app_php.png?etag=old' }), '/api/v2/apps/icon/php');
    assert.equal(getAppIconSrc({ key: 'php' }), '/api/v2/apps/icon/php');
});

function findComponent(node, tag) {
    if (node.tag === tag) return node;
    for (const child of node.children || []) {
        const component = findComponent(child, tag);
        if (component) return component;
    }
}

function mountAvatar(template, context) {
    const node = (type, text) => ({ type, text, children: [], props: {} });
    const renderer = Vue.createRenderer({
        createElement: (type) => node(type),
        createText: (text) => node('#text', text),
        createComment: (text) => node('#comment', text),
        setText: (element, text) => {
            element.text = text;
        },
        setElementText: (element, text) => {
            element.children = [node('#text', text)];
        },
        parentNode: (element) => element.parent,
        nextSibling: () => null,
        patchProp: (element, key, _, value) => {
            element.props[key] = value;
        },
        insert(child, parent) {
            parent.children.push(child);
            child.parent = parent;
        },
        remove(child) {
            child.parent.children = child.parent.children.filter((item) => item !== child);
        },
    });
    const render = new Function('Vue', compile(template, { mode: 'function', prefixIdentifiers: true }).code)(Vue);
    const app = renderer.createApp({ setup: () => context, render });
    app.component('ElAvatar', ElAvatar);
    app.component('AppIcon', {
        props: ['appKey', 'icon', 'currentNode'],
        setup: () => ({ getAppIconUrl }),
        render: new Function(
            'Vue',
            compile(readView('app-store/installed/app/icon.vue').template.content, {
                mode: 'function',
                prefixIdentifiers: true,
            }).code,
        )(Vue),
    });
    app.component('SvgIcon', {
        props: ['iconName'],
        render() {
            return Vue.h('svg', { 'data-icon': this.iconName });
        },
    });
    const root = node('root');
    app.mount(root);
    return { root, app };
}

for (const view of [
    'app-store/apps/app/index.vue',
    'app-store/detail/index.vue',
    'app-store/installed/app/icon.vue',
    'home/app/index.vue',
]) {
    test(`${view}: actual avatar binding loads CDN once and displays the local fallback on error`, async () => {
        const avatar = findComponent(baseParse(readView(view).template.content), 'el-avatar');
        assert.ok(avatar, 'image entry must use an avatar with a local fallback');
        const { root, app } = mountAvatar(avatar.loc.source, {
            app: { key: 'php', icon: cdnURL },
            appKey: 'php',
            icon: cdnURL,
            currentNode: 'child-node',
            getAppIconUrl,
            getAppIconSrc: launcherResolver(),
        });
        const element = root.children[0];
        const image = element.children.find((child) => child.type === 'img');
        assert.equal(image.props.src, cdnURL);
        image.props.onError(new Event('error'));
        await Vue.nextTick();
        assert.equal(
            element.children.some((child) => child.type === 'img'),
            false,
        );
        assert.equal(element.children.find((child) => child.type === 'svg').props['data-icon'], 'p-appstore');
        await Vue.nextTick();
        assert.equal(
            element.children.some((child) => child.type === 'img'),
            false,
            'failed URL is not retried',
        );
        app.unmount();
    });
}

test('installed card passes the API icon URL through to the actual icon child component', () => {
    const icon = findComponent(baseParse(readView('app-store/installed/app/card.vue').template.content), 'AppIcon');
    const { root, app } = mountAvatar(icon.loc.source, {
        installed: { appKey: 'php', icon: cdnURL },
        currentNode: 'child-node',
    });
    const image = root.children[0].children[0].children.find((child) => child.type === 'img');
    assert.equal(image.props.src, cdnURL);
    app.unmount();
});
