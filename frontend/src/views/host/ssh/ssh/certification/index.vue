<template>
    <div>
        <DrawerPro v-model="drawerVisible" :header="$t('ssh.pubkey')" @close="handleClose" size="large">
            <div class="mb-4">
                <el-alert :closable="false">{{ $t('ssh.pubKeyHelper', [currentUser]) }}</el-alert>
            </div>
            <el-button type="primary" plain @click="onOpenDialog('create')">
                {{ $t('commons.button.create') }}
            </el-button>
            <el-button plain @click="onSync()">
                {{ $t('commons.button.sync') }}
            </el-button>
            <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                {{ $t('commons.button.delete') }}
            </el-button>
            <ComplexTable
                :pagination-config="paginationConfig"
                v-model:selects="selects"
                :data="data"
                @search="search"
                :heightDiff="370"
            >
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" />
                <el-table-column :label="$t('ssh.encryptionMode')" prop="encryptionMode" />
                <el-table-column :label="$t('commons.table.description')" prop="description" />
                <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" />
            </ComplexTable>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </DrawerPro>

        <DialogPro v-model="connOpen" :title="$t('ssh.pubkey')" size="small" :showClose="false">
            <el-descriptions class="margin-top" :column="1" border>
                <el-descriptions-item align="center" :label="$t('ssh.password')">
                    <div>
                        <span>{{ loadPassPhrase() }}</span>
                    </div>
                    <el-button
                        v-if="currentRow.passPhrase && Base64.decode(currentRow.passPhrase) !== '<UN-SET>'"
                        size="small"
                        icon="CopyDocument"
                        @click="onCopy(currentRow.passPhrase)"
                    >
                        {{ $t('commons.button.copy') }}
                    </el-button>
                </el-descriptions-item>
                <el-descriptions-item align="center" :label="$t('ssh.publicKey')">
                    <el-button-group size="small">
                        <el-button icon="CopyDocument" @click="onCopy(currentRow.publicKey)">
                            {{ $t('commons.button.copy') }}
                        </el-button>
                        <el-button icon="Download" @click="onDownload(currentRow, 'publicKey')">
                            {{ $t('commons.button.download') }}
                        </el-button>
                    </el-button-group>
                </el-descriptions-item>
                <el-descriptions-item align="center" :label="$t('ssh.privateKey')">
                    <el-button-group size="small">
                        <el-button icon="CopyDocument" @click="onCopy(currentRow.privateKey)">
                            {{ $t('commons.button.copy') }}
                        </el-button>
                        <el-button icon="Download" @click="onDownload(currentRow, 'privateKey')">
                            {{ $t('commons.button.download') }}
                        </el-button>
                    </el-button-group>
                </el-descriptions-item>
            </el-descriptions>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="connOpen = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>
        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="forceDelete" :label="$t('website.forceDelete')" />
                        <span class="input-help">
                            {{ $t('website.forceDeleteHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <Operate ref="dialogRef" @search="search" />
    </div>
</template>
<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { deleteCert, searchCert, syncCert } from '@/api/modules/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import Operate from '@/views/host/ssh/ssh/certification/operate/index.vue';
import { copyText } from '@/utils/util';
import { Base64 } from 'js-base64';
import { reactive, ref } from 'vue';

const loading = ref();
const drawerVisible = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'login-log-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('login-log-page-size')) || 20,
    total: 0,
    small: true,
});

const forceDelete = ref();
const operateIDs = ref();
const opRef = ref();

const currentRow = ref();
const connOpen = ref();
const currentUser = ref();

const acceptParams = async (user: string): Promise<void> => {
    search();
    currentUser.value = user || 'root';
    drawerVisible.value = true;
};

const loadPassPhrase = () => {
    if (currentRow.value.passPhrase === '') {
        return '-';
    }
    let pass = Base64.decode(currentRow.value.passPhrase);
    return pass === '<UN-SET>' ? i18n.global.t('ssh.unSyncPass') : '';
};

const onCopy = async (content: string) => {
    content = Base64.decode(content);
    copyText(content);
};

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Host.RootCertInfo> = {
        mode: 'generate',
        encryptionMode: 'ed25519',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onDownload = async (row: Host.RootCertInfo, type: string) => {
    let name = row.name;
    let content;
    if (type === 'publicKey') {
        name = row.name + '.pub';
        content = Base64.decode(row.publicKey);
    } else {
        content = Base64.decode(row.privateKey);
    }
    const downloadUrl = window.URL.createObjectURL(new Blob([content], { type: 'application/octet-stream' }));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = name;
    const event = new MouseEvent('click');
    a.dispatchEvent(event);

    setTimeout(() => {
        document.body.removeChild(a);
        window.URL.revokeObjectURL(downloadUrl);
    }, 100);
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchCert(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSync = async () => {
    ElMessageBox.confirm(i18n.global.t('ssh.syncHelper'), i18n.global.t('commons.button.sync'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await syncCert()
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onDelete = async (row: Host.RootCertInfo | null) => {
    let names = [];
    let ids = [];
    forceDelete.value = false;
    if (row) {
        ids = [row.id];
        names = [row.name + ' - ' + row.encryptionMode];
    } else {
        for (const item of selects.value) {
            names.push(item.name + ' - ' + item.encryptionMode);
            ids.push(item.id);
        }
    }
    operateIDs.value = ids;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('menu.cronjob'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    await deleteCert(operateIDs.value, forceDelete.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisible.value = false;
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Host.RootCertInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.view'),
        click: (row: Host.RootCertInfo) => {
            currentRow.value = row;
            connOpen.value = true;
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RootCertInfo) => {
            onDelete(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.marginTop {
    margin-top: 10px;
}
</style>
