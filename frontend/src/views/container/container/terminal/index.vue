<template>
    <el-drawer v-model="terminalVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('container.containerTerminal')" :back="handleClose" />
        </template>
        <el-form ref="formRef" :model="form" label-position="top">
            <el-row :gutter="20" type="flex" justify="center">
                <el-col :span="12">
                    <el-form-item label="User" prop="user">
                        <el-input clearable v-model="form.user" />
                        <span class="input-help">{{ $t('container.emptyUser') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item v-if="form.isCustom" label="Command" prop="command" :rules="Rules.requiredInput">
                        <el-checkbox style="width: 100px" border v-model="form.isCustom" @change="onChangeCommand">
                            {{ $t('container.custom') }}
                        </el-checkbox>
                        <el-input style="width: calc(100% - 100px)" clearable v-model="form.command" />
                    </el-form-item>
                    <el-form-item v-if="!form.isCustom" label="Command" prop="command" :rules="Rules.requiredSelect">
                        <el-checkbox style="width: 100px" border v-model="form.isCustom" @change="onChangeCommand">
                            {{ $t('container.custom') }}
                        </el-checkbox>
                        <el-select style="width: calc(100% - 100px)" filterable clearable v-model="form.command">
                            <el-option value="/bin/ash" label="/bin/ash" />
                            <el-option value="/bin/bash" label="/bin/bash" />
                            <el-option value="/bin/sh" label="/bin/sh" />
                        </el-select>
                    </el-form-item>
                </el-col>
            </el-row>
            <div style="height: calc(100vh - 290px)" :id="'terminal-exec'"></div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="terminalVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" v-if="!terminalOpen" @click="initTerm(formRef)">
                    {{ $t('commons.button.conn') }}
                </el-button>
                <el-button type="primary" v-else @click="handleClose()">{{ $t('commons.button.disconn') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { ElForm, FormInstance } from 'element-plus';
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { Base64 } from 'js-base64';
import 'xterm/css/xterm.css';
import { FitAddon } from 'xterm-addon-fit';
import { Rules } from '@/global/form-rules';
import { isJson } from '@/utils/util';
import DrawerHeader from '@/components/drawer-header/index.vue';

const terminalVisiable = ref(false);
const terminalOpen = ref(false);
const fitAddon = new FitAddon();
let terminalSocket = ref(null) as unknown as WebSocket;
let term = ref(null) as unknown as Terminal;
const loading = ref(true);
const runRealTerminal = () => {
    loading.value = false;
};
const form = reactive({
    isCustom: false,
    command: '',
    user: '',
    containerID: '',
});
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    containerID: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisiable.value = true;
    form.containerID = params.containerID;
    form.isCustom = false;
    form.user = '';
    form.command = '/bin/bash';
    terminalOpen.value = false;
    window.addEventListener('resize', changeTerminalSize);
};

const onChangeCommand = async () => {
    console.log('addqwd');
    form.command = '';
};

const onWSReceive = (message: any) => {
    if (!isJson(message.data)) {
        return;
    }
    const data = JSON.parse(message.data);
    term.element && term.focus();
    term.write(data.Data);
};

const errorRealTerminal = (ex: any) => {
    let message = ex.message;
    if (!message) message = 'disconnected';
    term.write(`\x1b[31m${message}\x1b[m\r\n`);
};

const closeRealTerminal = (ev: CloseEvent) => {
    term.write(ev.reason);
};

const initTerm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let href = window.location.href;
        let ipLocal = href.split('//')[1].split('/')[0];
        terminalOpen.value = true;
        let ifm = document.getElementById('terminal-exec') as HTMLInputElement | null;
        term = new Terminal({
            lineHeight: 1.2,
            fontSize: 12,
            fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
            theme: {
                background: '#000000',
            },
            cursorBlink: true,
            cursorStyle: 'underline',
            scrollback: 100,
            tabStopWidth: 4,
        });
        if (ifm) {
            term.open(ifm);
            terminalSocket = new WebSocket(
                `ws://${ipLocal}/api/v1/containers/exec?containerid=${form.containerID}&cols=${term.cols}&rows=${term.rows}&user=${form.user}&command=${form.command}`,
            );
            terminalSocket.onopen = runRealTerminal;
            terminalSocket.onmessage = onWSReceive;
            terminalSocket.onclose = closeRealTerminal;
            terminalSocket.onerror = errorRealTerminal;
            term.onData((data: any) => {
                if (isWsOpen()) {
                    terminalSocket.send(
                        JSON.stringify({
                            type: 'cmd',
                            cmd: Base64.encode(data),
                        }),
                    );
                }
            });
            term.loadAddon(new AttachAddon(terminalSocket));
            term.loadAddon(fitAddon);
            setTimeout(() => {
                fitAddon.fit();
                if (isWsOpen()) {
                    terminalSocket.send(
                        JSON.stringify({
                            type: 'resize',
                            cols: term.cols,
                            rows: term.rows,
                        }),
                    );
                }
            }, 30);
        }
    });
};

const fitTerm = () => {
    fitAddon.fit();
};

const isWsOpen = () => {
    const readyState = terminalSocket && terminalSocket.readyState;
    if (readyState) {
        return readyState === 1;
    }
    return false;
};

function handleClose() {
    terminalVisiable.value = false;
    terminalOpen.value = false;
    window.removeEventListener('resize', changeTerminalSize);
    if (isWsOpen()) {
        terminalSocket && terminalSocket.close();
        term.dispose();
    }
}

function changeTerminalSize() {
    fitTerm();
    const { cols, rows } = term;
    if (isWsOpen()) {
        terminalSocket.send(
            JSON.stringify({
                type: 'resize',
                cols: cols,
                rows: rows,
            }),
        );
    }
}
defineExpose({
    acceptParams,
});
</script>
