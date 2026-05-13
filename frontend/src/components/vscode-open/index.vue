<template>
    <DialogPro v-model="open" :title="$t('file.openWithVscode')" size="large" @close="handleClose">
        <div v-loading="loading" class="vscode-open space-y-4">
            <el-alert
                v-if="addForm.authMode === 'key'"
                :closable="false"
                :title="$t('file.vscodeHelper')"
                :description="$t('file.vscodeKeyHelper')"
                show-icon
                type="info"
            />

            <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(360px,0.9fr)]">
                <section class="vscode-open__panel">
                    <el-form ref="vscodeConnectInfoForm" :model="addForm" :rules="rules" label-position="top">
                        <el-form-item :label="$t('terminal.authMode')" prop="authMode" class="mb-5">
                            <el-radio-group v-model="addForm.authMode" @change="handleAuthModeChange">
                                <el-radio-button value="password">{{ $t('terminal.passwordMode') }}</el-radio-button>
                                <el-radio-button value="key">{{ $t('terminal.keyMode') }}</el-radio-button>
                            </el-radio-group>
                        </el-form-item>

                        <div class="grid gap-4 md:grid-cols-2">
                            <el-form-item :label="$t('terminal.ip')" prop="host">
                                <el-input v-model.trim="addForm.host" autocomplete="off" />
                            </el-form-item>

                            <el-form-item :label="$t('commons.table.port')" prop="port">
                                <el-input v-model.number="addForm.port" autocomplete="off" />
                            </el-form-item>
                        </div>

                        <el-form-item :label="$t('commons.login.username')" prop="username">
                            <el-input v-model.trim="addForm.username" autocomplete="off" />
                        </el-form-item>

                        <template v-if="addForm.authMode === 'key'">
                            <el-form-item :label="$t('file.vscodeSelectKey')" prop="certID">
                                <el-select
                                    v-model="addForm.certID"
                                    class="w-full"
                                    clearable
                                    filterable
                                    :loading="loading"
                                    :placeholder="$t('commons.msg.inputOrSelect')"
                                    @change="handleCertChange"
                                >
                                    <el-option
                                        v-for="item in certOptions"
                                        :key="item.id"
                                        :label="formatCertLabel(item)"
                                        :value="item.id"
                                    />
                                </el-select>
                            </el-form-item>

                            <el-form-item :label="$t('file.vscodeKeyPath')" prop="sshKeyPath">
                                <el-input
                                    v-model.trim="addForm.sshKeyPath"
                                    :placeholder="$t('file.vscodeKeyPathPlaceholder')"
                                    autocomplete="off"
                                />
                            </el-form-item>
                        </template>
                    </el-form>
                </section>

                <section class="vscode-open__panel">
                    <div v-if="addForm.authMode === 'key'" class="space-y-4">
                        <div class="flex items-center justify-between gap-2">
                            <div>
                                <div class="vscode-open__title">
                                    {{ $t('file.vscodeScriptPreview') }}
                                </div>
                                <p class="vscode-open__tip">
                                    {{ $t('file.vscodeScriptPreviewHint') }}
                                </p>
                            </div>
                            <div class="flex flex-wrap justify-end gap-2">
                                <el-button
                                    :disabled="!canGenerate"
                                    plain
                                    @click="showScriptPreview = !showScriptPreview"
                                >
                                    {{
                                        showScriptPreview ? $t('commons.button.collapse') : $t('commons.button.preview')
                                    }}
                                </el-button>
                                <el-button :disabled="!canGenerate" plain @click="copySetupScript">
                                    {{ $t('file.vscodeCopyConfig') }}
                                </el-button>
                            </div>
                        </div>

                        <pre
                            v-if="showScriptPreview"
                            class="vscode-open__script max-h-[430px] overflow-auto p-3 text-xs leading-6"
                            >{{ scriptText }}</pre
                        >
                    </div>

                    <div v-else class="flex h-full min-h-[320px] flex-col justify-between gap-4">
                        <div class="space-y-3">
                            <div class="vscode-open__title">
                                {{ $t('file.vscodePasswordModeTitle') }}
                            </div>
                            <p class="vscode-open__tip">
                                {{ $t('file.vscodePasswordModeHint') }}
                            </p>
                            <div class="vscode-open__steps space-y-2 p-3">
                                <div class="flex items-center gap-2 text-sm">
                                    <span class="vscode-open__step-index">1</span>
                                    {{ $t('file.vscodePasswordStep1') }}
                                </div>
                                <div class="flex items-center gap-2 text-sm">
                                    <span class="vscode-open__step-index">2</span>
                                    {{ $t('file.vscodePasswordStep2') }}
                                </div>
                                <div class="flex items-center gap-2 text-sm">
                                    <span class="vscode-open__step-index">3</span>
                                    {{ $t('file.vscodePasswordStep3') }}
                                </div>
                            </div>
                        </div>
                    </div>
                </section>
            </div>
        </div>

        <template #footer>
            <span class="dialog-footer flex flex-wrap items-center justify-end gap-2">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button
                    v-if="addForm.authMode === 'key'"
                    :disabled="!selectedCert"
                    plain
                    @click="downloadPrivateKey"
                >
                    {{ $t('commons.button.download') }}
                </el-button>
                <el-button v-if="addForm.authMode === 'key'" :disabled="!canGenerate" @click="copySetupScript">
                    {{ $t('file.vscodeCopyConfig') }}
                </el-button>
                <el-button :disabled="!canGenerate" type="primary" @click="submit">
                    {{ $t('file.openWithVscode') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import type { FormInstance, FormItemRule, FormRules } from 'element-plus';
import { Base64 } from 'js-base64';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getSSHInfo, searchCert } from '@/api/modules/host';
import { loadLocalConn } from '@/api/modules/terminal';
import { Host } from '@/api/interface/host';
import { Rules } from '@/global/form-rules';
import { copyText } from '@/utils/clipboard';
import { MsgError } from '@/utils/message';

const { currentNode, currentNodeAddr } = useGlobalStore();
const open = ref(false);
const loading = ref(false);
const showScriptPreview = ref(false);
const vscodeConnectInfoForm = ref<FormInstance>();
const certOptions = ref<Host.RootCertInfo[]>([]);

interface DialogProps {
    path: string;
}

const STORAGE_KEY = 'VscodeConnectInfo';
type AuthMode = 'key' | 'password';

const normalizeAuthMode = (value: unknown): AuthMode => (value === 'key' ? 'key' : 'password');

const defaultForm = () => ({
    authMode: 'password' as AuthMode,
    host: '',
    port: 22,
    username: 'root',
    certID: '' as number | '',
    sshKeyPath: '',
    path: '',
});

const addForm = reactive(defaultForm());

const getStorageKey = () => `${STORAGE_KEY}:${currentNode.value}:${currentNodeAddr.value || 'local'}`;

const isKeyMode = computed(() => addForm.authMode === 'key');
const selectedCert = computed(() => certOptions.value.find((item) => String(item.id) === String(addForm.certID)));

const sshConfigValue = (value: string) =>
    `"${value.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\$/g, '\\$')}"`;

const normalizeHost = (value: string) => value.trim().replace(/^\[(.*)\]$/, '$1');

const formatHostForRemote = (value: string) => {
    const trimmed = normalizeHost(value);
    if (trimmed.includes(':') && !trimmed.startsWith('[')) {
        return `[${trimmed}]`;
    }
    return trimmed;
};

const getKeyPathFileName = () => {
    const keyPath = addForm.sshKeyPath.trim();
    const normalizedPath = keyPath.replace(/\\/g, '/').replace(/\/+$/, '');
    return normalizedPath.split('/').pop() || selectedCert.value?.name || 'ssh-private-key';
};

const getSafeKeyFileName = (name = 'ssh-private-key') =>
    name
        .trim()
        .replace(/[/\\]/g, '-')
        .replace(/\s+/g, '-')
        .replace(/[^a-zA-Z0-9._-]+/g, '-')
        .replace(/-+/g, '-')
        .replace(/^[.-]+|[.-]+$/g, '') || 'ssh-private-key';

const buildSshAlias = () => {
    const host = normalizeHost(addForm.host);
    const port = Number(addForm.port) || 22;
    const username = addForm.username.trim() || 'root';
    const certName = selectedCert.value?.name?.trim() || 'key';
    const certID = selectedCert.value?.id ?? 0;
    return `1panel-${host}-${port}-${username}-${certName}-${certID}`
        .replace(/[^a-zA-Z0-9._-]+/g, '-')
        .replace(/-+/g, '-');
};

const buildVscodeUrl = () => {
    const remoteTarget = isKeyMode.value
        ? buildSshAlias()
        : `${addForm.username.trim() || 'root'}@${formatHostForRemote(addForm.host)}:${Number(addForm.port) || 22}`;
    const remotePath = encodeURI(addForm.path);
    return `vscode://vscode-remote/ssh-remote+${encodeURIComponent(remoteTarget)}${remotePath}?windowId=_blank`;
};

const buildSetupScript = () => {
    const host = normalizeHost(addForm.host);
    const port = Number(addForm.port) || 22;
    const username = addForm.username.trim() || 'root';
    const keyPath = addForm.sshKeyPath.trim();
    const alias = buildSshAlias();
    const managedBegin = `# 1Panel VS Code managed block: ${alias}`;
    const managedEnd = `# 1Panel VS Code managed block end: ${alias}`;

    return [
        '#!/usr/bin/env bash',
        'set -euo pipefail',
        '',
        '# Run this script on the computer running VS Code.',
        `SSH_CONFIG="$HOME/.ssh/config"`,
        `KEY_PATH=${sshConfigValue(keyPath)}`,
        `MANAGED_BEGIN=${sshConfigValue(managedBegin)}`,
        `MANAGED_END=${sshConfigValue(managedEnd)}`,
        '',
        'mkdir -p "$(dirname "$SSH_CONFIG")"',
        'chmod 700 "$(dirname "$SSH_CONFIG")" || true',
        'touch "$SSH_CONFIG"',
        '',
        'tmp_file="$(mktemp)"',
        'awk -v begin="$MANAGED_BEGIN" -v end="$MANAGED_END" \'',
        '  $0 == begin { skip = 1; next }',
        '  $0 == end { skip = 0; next }',
        '  !skip { print }',
        '\' "$SSH_CONFIG" > "$tmp_file"',
        '',
        'cat <<EOF >> "$tmp_file"',
        managedBegin,
        `Host ${alias}`,
        `    HostName ${sshConfigValue(host)}`,
        `    User ${sshConfigValue(username)}`,
        `    Port ${port}`,
        `    IdentityFile ${sshConfigValue(keyPath)}`,
        '    IdentitiesOnly yes',
        managedEnd,
        'EOF',
        '',
        'mv "$tmp_file" "$SSH_CONFIG"',
        'chmod 600 "$SSH_CONFIG"',
        'if [ ! -f "$KEY_PATH" ]; then',
        '    echo "1Panel: private key not found: $KEY_PATH" >&2',
        '    echo "1Panel: download the private key and save it to this exact path, then rerun this script." >&2',
        '    exit 1',
        'fi',
        'chmod 600 "$KEY_PATH"',
        `echo "1Panel: SSH config ready for ${alias}"`,
    ].join('\n');
};

const scriptText = computed(() =>
    addForm.authMode === 'key' && addForm.path && canGenerate.value ? buildSetupScript() : '',
);

const canGenerate = computed(() => {
    if (!addForm.host.trim() || !addForm.username.trim() || !addForm.port) {
        return false;
    }
    if (!isKeyMode.value) {
        return true;
    }
    return Boolean(selectedCert.value && addForm.sshKeyPath.trim());
});

const requiredKeyValidator = (_rule: FormItemRule, value: unknown, callback: (error?: Error) => void) => {
    if (!isKeyMode.value) {
        callback();
        return;
    }
    if (value === '' || typeof value === 'undefined' || value === null) {
        callback(new Error(String(i18n.global.t('commons.rule.requiredSelect'))));
        return;
    }
    callback();
};

const requiredKeyPathValidator = (_rule: FormItemRule, value: unknown, callback: (error?: Error) => void) => {
    if (!isKeyMode.value) {
        callback();
        return;
    }
    if (value === '' || typeof value === 'undefined' || value === null) {
        callback(new Error(String(i18n.global.t('commons.rule.requiredInput'))));
        return;
    }
    callback();
};

const rules = reactive<FormRules>({
    authMode: [Rules.requiredSelect],
    host: [Rules.requiredInput, Rules.ipV4V6OrDomain],
    port: [Rules.requiredInput, Rules.port],
    username: [Rules.requiredInput],
    certID: [{ validator: requiredKeyValidator, trigger: 'change' }],
    sshKeyPath: [{ validator: requiredKeyPathValidator, trigger: 'blur' }],
});

const emit = defineEmits(['close']);

const saveDraft = () => {
    localStorage.setItem(
        getStorageKey(),
        JSON.stringify({
            authMode: addForm.authMode,
            host: addForm.host,
            port: addForm.port,
            username: addForm.username,
            certID: addForm.certID,
            sshKeyPath: addForm.sshKeyPath,
        }),
    );
};

const restoreDraft = () => {
    const storedInfo = localStorage.getItem(getStorageKey());
    if (!storedInfo) return;

    try {
        const parsed = JSON.parse(storedInfo);
        addForm.authMode = normalizeAuthMode(parsed.authMode);
        addForm.host = parsed.host?.trim() || addForm.host;
        addForm.port = parsed.port || addForm.port;
        addForm.username = parsed.username?.trim() || addForm.username;
        addForm.certID = parsed.certID ?? '';
        addForm.sshKeyPath = parsed.sshKeyPath?.trim() || '';
    } catch {
        localStorage.removeItem(getStorageKey());
    }
};

const loadConnectionInfo = async () => {
    if (currentNode.value === 'local') {
        try {
            const res = await loadLocalConn();
            if (res.data) {
                addForm.host = res.data.addr || currentNodeAddr.value || '127.0.0.1';
                addForm.port = Number(res.data.port) || 22;
                addForm.username = res.data.user || 'root';
                return;
            }
        } catch {}
    }

    try {
        const res = await getSSHInfo();
        if (res.data) {
            addForm.port = Number(res.data.port) || 22;
            addForm.username = res.data.currentUser || 'root';
        }
    } catch {}

    addForm.host = currentNodeAddr.value || addForm.host || '127.0.0.1';
};

const loadCertOptions = async () => {
    const res = await searchCert({ page: 1, pageSize: 500 });
    certOptions.value = res.data?.items || [];
};

const loadDialogData = async () => {
    loading.value = true;
    await Promise.allSettled([loadConnectionInfo(), loadCertOptions()]);
    restoreDraft();
    if (addForm.authMode === 'key' && addForm.certID && !selectedCert.value) {
        addForm.certID = '';
    }
    loading.value = false;
};

const formatCertLabel = (item: Host.RootCertInfo) => {
    const desc = item.description?.trim();
    return desc ? `${item.name} (${item.encryptionMode}) - ${desc}` : `${item.name} (${item.encryptionMode})`;
};

const handleClose = () => {
    saveDraft();
    open.value = false;
    showScriptPreview.value = false;
    vscodeConnectInfoForm.value?.resetFields();
    Object.assign(addForm, defaultForm());
    certOptions.value = [];
    emit('close', false);
};

const handleAuthModeChange = () => {
    showScriptPreview.value = false;
};

const handleCertChange = () => {
    if (addForm.sshKeyPath.trim() || !selectedCert.value?.name) {
        return;
    }
    addForm.sshKeyPath = `~/.ssh/${getSafeKeyFileName(selectedCert.value.name)}`;
};

const copySetupScript = async () => {
    if (addForm.authMode !== 'key') return;
    if (!canGenerate.value) {
        MsgError(String(i18n.global.t('commons.msg.inputOrSelect')));
        return;
    }
    await copyText(buildSetupScript());
};

const downloadPrivateKey = () => {
    if (!selectedCert.value?.privateKey) {
        MsgError(String(i18n.global.t('commons.msg.noneData')));
        return;
    }
    const content = Base64.decode(selectedCert.value.privateKey);
    const fileName = getKeyPathFileName();
    const downloadUrl = window.URL.createObjectURL(new Blob([content], { type: 'application/octet-stream' }));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = fileName;
    a.dispatchEvent(new MouseEvent('click'));

    setTimeout(() => {
        window.URL.revokeObjectURL(downloadUrl);
    }, 100);
};

const submit = async () => {
    if (!vscodeConnectInfoForm.value) return;

    await vscodeConnectInfoForm.value.validate((valid) => {
        if (!valid || !canGenerate.value) return;

        saveDraft();
        open.value = false;
        window.open(buildVscodeUrl(), '_blank', 'noopener,noreferrer');
    });
};

const acceptParams = async (params: DialogProps): Promise<void> => {
    Object.assign(addForm, defaultForm());
    certOptions.value = [];
    showScriptPreview.value = false;
    addForm.path = params.path;
    open.value = true;
    await loadDialogData();
};

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.vscode-open {
    :deep(.el-descriptions__label) {
        width: 120px;
    }
}

.vscode-open__panel {
    padding: 16px;
    border: 1px solid var(--el-border-color-light);
    border-radius: var(--el-border-radius-base);
    background: var(--el-bg-color);
}

.vscode-open__title {
    color: var(--el-text-color-primary);
    font-size: 15px;
    font-weight: 600;
    line-height: 24px;
}

.vscode-open__tip {
    margin-top: 4px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 22px;
}

.vscode-open__script {
    border: 1px solid var(--el-border-color-light);
    border-radius: var(--el-border-radius-base);
    color: var(--el-text-color-primary);
    background: var(--el-fill-color-lighter);
}

.vscode-open__steps {
    border: 1px dashed var(--el-border-color);
    border-radius: var(--el-border-radius-base);
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-lighter);
}

.vscode-open__step-index {
    display: inline-flex;
    width: 22px;
    height: 22px;
    flex: none;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    font-size: 12px;
    font-weight: 600;
}

@media (max-width: 768px) {
    .vscode-open {
        :deep(.el-descriptions__table) {
            display: block;
        }

        :deep(.el-descriptions__body),
        :deep(.el-descriptions__table tbody),
        :deep(.el-descriptions__table tr),
        :deep(.el-descriptions__table th),
        :deep(.el-descriptions__table td) {
            display: block;
            width: 100% !important;
        }

        :deep(.el-descriptions__label) {
            width: 100%;
        }
    }

    .vscode-open__panel {
        padding: 12px;
    }
}
</style>
