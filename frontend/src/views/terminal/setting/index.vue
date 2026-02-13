<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('container.setting')" :divider="true">
            <template #main>
                <el-form :model="form" label-position="left" label-width="150px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                            <el-form-item :label="$t('terminal.lineHeight')">
                                <el-input-number
                                    class="formInput"
                                    :min="1"
                                    :max="2.0"
                                    :precision="1"
                                    :step="0.1"
                                    v-model="form.lineHeight"
                                    @change="changeItem()"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('terminal.letterSpacing')">
                                <el-input-number
                                    class="formInput"
                                    :min="0"
                                    :max="3.5"
                                    :precision="1"
                                    :step="0.5"
                                    v-model="form.letterSpacing"
                                    @change="changeItem()"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('terminal.fontSize')">
                                <el-input-number
                                    class="formInput"
                                    :step="1"
                                    :min="12"
                                    :max="20"
                                    v-model="form.fontSize"
                                    @change="changeItem()"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('terminal.fontFamily')">
                                <el-input
                                    class="formInput"
                                    v-model="form.fontFamily"
                                    :placeholder="$t('terminal.fontFamily')"
                                    @change="changeItem()"
                                />
                                <span class="input-help">{{ $t('terminal.fontFamilyHelper') }}</span>
                            </el-form-item>

                            <el-form-item>
                                <div class="terminal" ref="terminalElement"></div>
                            </el-form-item>

                            <el-form-item :label="$t('terminal.cursorBlink')">
                                <el-switch
                                    v-model="form.cursorBlink"
                                    active-value="Enable"
                                    inactive-value="Disable"
                                    @change="changeItem()"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('terminal.cursorStyle')">
                                <el-select class="formInput" v-model="form.cursorStyle" @change="changeItem()">
                                    <el-option value="block" :label="$t('terminal.cursorBlock')" />
                                    <el-option value="underline" :label="$t('terminal.cursorUnderline')" />
                                    <el-option value="bar" :label="$t('terminal.cursorBar')" />
                                </el-select>
                            </el-form-item>
                            <el-form-item :label="$t('terminal.scrollback')">
                                <el-input-number
                                    class="formInput"
                                    :step="50"
                                    :min="0"
                                    :max="10000"
                                    v-model="form.scrollback"
                                    @change="changeItem()"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('terminal.scrollSensitivity')">
                                <el-input-number
                                    class="formInput"
                                    :step="1"
                                    :min="0"
                                    :max="16"
                                    v-model="form.scrollSensitivity"
                                    @change="changeItem()"
                                />
                            </el-form-item>
                            <el-form-item>
                                <el-button @click="onSetDefault()" plain>
                                    {{ $t('commons.button.setDefault') }}
                                </el-button>
                                <el-button @click="search(true)" plain>{{ $t('commons.button.reset') }}</el-button>
                                <el-button @click="onSave" type="primary">{{ $t('commons.button.save') }}</el-button>
                            </el-form-item>
                            <el-divider border-style="dashed" />
                            <el-form-item :label="$t('terminal.defaultConn')">
                                <el-switch v-model="form.showDefaultConn" @change="changeShow" />
                            </el-form-item>
                            <el-form-item :label="$t('xpack.node.connInfo')">
                                <el-input disabled v-model="form.defaultConn">
                                    <template #append>
                                        <el-button @click="dialogRef.acceptParams(false)" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <OperateDialog @search="loadConnShow" ref="dialogRef" />

        <OpDialog ref="opRef" @search="search" @cancel="loadConnShow" @submit="submitChangeShow">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" v-if="!form.showDefaultConn" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="resetConn" :label="$t('terminal.withReset')" />
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import { ElForm } from 'element-plus';
import { getTerminalInfo, UpdateTerminalInfo } from '@/api/modules/setting';
import { Terminal } from '@xterm/xterm';
import OperateDialog from '@/views/terminal/setting/default_conn/index.vue';
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { TerminalStore } from '@/store';
import { loadLocalConn, updateLocalConn } from '@/api/modules/terminal';

const loading = ref(false);
const terminalStore = TerminalStore();
const dialogRef = ref();

const terminalElement = ref<HTMLDivElement | null>(null);
const fitAddon = new FitAddon();
const term = ref();

const form = reactive({
    lineHeight: 1.2,
    letterSpacing: 1.2,
    fontSize: 12,
    fontFamily: '',
    cursorBlink: 'Enable',
    cursorStyle: 'underline',
    scrollback: 1000,
    scrollSensitivity: 10,

    showDefaultConn: false,
    defaultConn: '',
});

const resetConn = ref(false);
const opRef = ref();

const acceptParams = () => {
    search(true);
    loadConnShow();
    iniTerm();
};

const search = async (withReset?: boolean) => {
    loading.value = true;
    await getTerminalInfo()
        .then((res) => {
            loading.value = false;
            form.lineHeight = Number(res.data.lineHeight);
            form.letterSpacing = Number(res.data.letterSpacing);
            form.fontSize = Number(res.data.fontSize);
            form.fontFamily = res.data.fontFamily || '';
            form.cursorBlink = res.data.cursorBlink;
            form.cursorStyle = res.data.cursorStyle;
            form.scrollback = Number(res.data.scrollback);
            form.scrollSensitivity = Number(res.data.scrollSensitivity);

            // 同步到 store，确保已打开的终端也能使用新字体
            terminalStore.setFontFamily(res.data.fontFamily || '');

            if (withReset) {
                changeItem();
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadConnShow = async () => {
    await loadLocalConn().then((res) => {
        form.showDefaultConn = res.data.localSSHConnShow === 'Enable';
        if (res.data.addr && res.data.port && res.data.user) {
            form.defaultConn = res.data.user + '@' + res.data.addr + ':' + res.data.port;
        } else {
            form.defaultConn = '-';
        }
        resetConn.value = false;
    });
};

const changeShow = async () => {
    let op = form.showDefaultConn ? i18n.global.t('xpack.waf.allow') : i18n.global.t('xpack.waf.deny');
    opRef.value.acceptParams({
        title: i18n.global.t('terminal.defaultConn'),
        names: [],
        msg: i18n.global.t('terminal.defaultConnHelper', [op]),
        api: null,
        params: {},
    });
};
const submitChangeShow = async () => {
    loading.value = true;
    await updateLocalConn({
        withReset: resetConn.value,
        defaultConn: form.showDefaultConn ? 'Enable' : 'Disable',
    })
        .then(() => {
            loading.value = false;
            loadConnShow();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const iniTerm = () => {
    const defaultFontFamily = "Monaco, Menlo, Consolas, 'Courier New', monospace";
    const fontFamily = form.fontFamily || defaultFontFamily;

    term.value = new Terminal({
        lineHeight: 1.2,
        fontSize: 12,
        fontFamily: fontFamily,
        theme: {
            background: '#000000',
        },
        cursorBlink: true,
        cursorStyle: 'block',
        scrollback: 1000,
        scrollSensitivity: 15,
    });
    term.value.open(terminalElement.value);
    term.value.loadAddon(fitAddon);
    term.value.write('the first line \r\nthe second line');
    fitAddon.fit();
};

const changeItem = () => {
    const defaultFontFamily = "Monaco, Menlo, Consolas, 'Courier New', monospace";
    const fontFamily = form.fontFamily || defaultFontFamily;

    term.value.options.lineHeight = form.lineHeight;
    term.value.options.letterSpacing = form.letterSpacing;
    term.value.options.fontSize = form.fontSize;
    term.value.options.fontFamily = fontFamily;
    term.value.options.cursorBlink = form.cursorBlink === 'Enable';
    term.value.options.cursorStyle = form.cursorStyle;
    term.value.options.scrollback = form.scrollback;
    term.value.options.scrollSensitivity = form.scrollSensitivity;

    fitAddon.fit();
};

const onSetDefault = () => {
    form.lineHeight = 1.2;
    form.letterSpacing = 0;
    form.fontSize = 12;
    form.fontFamily = '';
    form.cursorBlink = 'Enable';
    form.cursorStyle = 'block';
    form.scrollback = 1000;
    form.scrollSensitivity = 6;

    changeItem();
};

const onSave = () => {
    ElMessageBox.confirm(i18n.global.t('terminal.saveHelper'), i18n.global.t('container.setting'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let param = {
            lineHeight: form.lineHeight + '',
            letterSpacing: form.letterSpacing + '',
            fontSize: form.fontSize + '',
            fontFamily: form.fontFamily,
            cursorBlink: form.cursorBlink,
            cursorStyle: form.cursorStyle,
            scrollback: form.scrollback + '',
            scrollSensitivity: form.scrollSensitivity + '',
        };
        await UpdateTerminalInfo(param)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                terminalStore.setLineHeight(form.lineHeight);
                terminalStore.setLetterSpacing(form.letterSpacing);
                terminalStore.setFontSize(form.fontSize);
                terminalStore.setFontFamily(form.fontFamily);
                terminalStore.setCursorBlink(form.cursorBlink);
                terminalStore.setCursorStyle(form.cursorStyle);
                terminalStore.setScrollback(form.scrollback);
                terminalStore.setScrollSensitivity(form.scrollSensitivity);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="css" scoped>
.formInput {
    width: 100%;
}
.terminal {
    width: 100%;
    height: 100px;
}
</style>
