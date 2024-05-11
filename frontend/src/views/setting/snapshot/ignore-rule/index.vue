<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.ignoreRule')" :back="handleClose" />
            </template>

            <el-alert :closable="false" type="warning">{{ $t('setting.ignoreHelper') }}</el-alert>
            <el-row type="flex" justify="center" v-loading="loading" class="mt-2">
                <el-col :span="22">
                    <el-form ref="formRef" :model="form" :rules="rules">
                        <el-form-item prop="tmpRule">
                            <div class="w-full">
                                <el-input
                                    v-model="form.tmpRule"
                                    :rows="5"
                                    style="width: calc(100% - 50px)"
                                    type="textarea"
                                    :placeholder="$t('setting.ignoreHelper1')"
                                />
                                <FileList @choose="loadDir" :path="baseDir" :isAll="true"></FileList>
                            </div>
                        </el-form-item>
                    </el-form>

                    <el-button :disabled="form.tmpRule === ''" @click="handleAdd(formRef)">
                        {{ $t('xpack.tamper.addRule') }}
                    </el-button>

                    <el-table :data="tableList">
                        <el-table-column prop="value" />
                        <el-table-column min-width="18">
                            <template #default="scope">
                                <el-button link type="primary" @click="handleDelete(scope.$index)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave()">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import FileList from '@/components/file-list/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { FormInstance } from 'element-plus';
import { getSettingInfo, loadBaseDir, updateSetting } from '@/api/modules/setting';

const loading = ref();
const baseDir = ref();
const drawerVisible = ref(false);
const tableList = ref();

const form = reactive({
    tmpRule: '',
});
const formRef = ref<FormInstance>();
const rules = reactive({
    tmpRule: [{ validator: checkData, trigger: 'blur' }],
});
function checkData(rule: any, value: any, callback: any) {
    if (form.tmpRule !== '') {
        const reg = /^[^\\\"'|<>?]{1,128}$/;
        let items = value.split('\n');
        for (const item of items) {
            if (item.indexOf(' ') !== -1) {
                callback(new Error(i18n.global.t('setting.noSpace')));
            }
            if (!reg.test(item) && value !== '') {
                callback(new Error(i18n.global.t('commons.rule.linuxName', ['\\:?\'"<>|'])));
            } else {
                callback();
            }
        }
    }
    callback();
}

const acceptParams = async (): Promise<void> => {
    loadPath();
    const res = await getSettingInfo();
    tableList.value = [];
    let items = res.data.snapshotIgnore.split(',');
    for (const item of items) {
        tableList.value.push({ value: item });
    }
    drawerVisible.value = true;
};

const loadPath = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
};

const loadDir = async (path: string) => {
    form.tmpRule += path + '\n';
};

const handleAdd = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let itemData = form.tmpRule.split('\n');
        for (const item of itemData) {
            if (item) {
                tableList.value.push({ value: item });
            }
        }
        form.tmpRule = '';
    });
};
const handleDelete = (index: number) => {
    tableList.value.splice(index, 1);
};

const onSave = async () => {
    let list = [];
    for (const item of tableList.value) {
        list.push(item.value);
    }
    await updateSetting({ key: 'SnapshotIgnore', value: list.join(',') })
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loading.value = false;
            drawerVisible.value = false;
            return;
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
