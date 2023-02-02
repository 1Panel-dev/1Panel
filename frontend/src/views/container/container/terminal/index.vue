<template>
    <el-dialog
        v-model="terminalVisiable"
        :destroy-on-close="true"
        @close="onClose"
        :close-on-click-modal="false"
        width="70%"
    >
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.containerTerminal') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item label="User" prop="user">
                <el-input style="width: 30%" clearable v-model="form.user" />
                <span class="input-help">{{ $t('container.emptyUser') }}</span>
            </el-form-item>
            <el-form-item :label="$t('container.custom')" prop="custom">
                <el-switch v-model="form.isCustom" @change="onChangeCommand" />
            </el-form-item>
            <el-form-item v-if="form.isCustom" label="Command" prop="command" :rules="Rules.requiredInput">
                <el-input style="width: 30%" clearable v-model="form.command" />
            </el-form-item>
            <el-form-item v-if="!form.isCustom" label="Command" prop="command" :rules="Rules.requiredSelect">
                <el-select style="width: 30%" allow-create filterable clearable v-model="form.command">
                    <el-option value="/bin/ash" label="/bin/ash" />
                    <el-option value="/bin/bash" label="/bin/bash" />
                    <el-option value="/bin/sh" label="/bin/sh" />
                </el-select>
            </el-form-item>
            <el-form-item>
                <el-button v-if="!terminalOpen" @click="initTerm(formRef)">{{ $t('commons.button.conn') }}</el-button>
                <el-button v-else @click="onClose()">{{ $t('commons.button.disconn') }}</el-button>
            </el-form-item>
        </el-form>
        <div :id="'terminal-exec'"></div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="terminalVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </el-dialog>
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

function isJson(str: string) {
    try {
        if (typeof JSON.parse(str) === 'object') {
            return true;
        }
    } catch {
        return false;
    }
}

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

function onClose() {
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
