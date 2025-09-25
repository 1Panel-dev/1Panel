<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.release')" @close="handleClose" size="large">
        <div class="note" v-loading="loading">
            <el-form ref="formRef" :model="form" :rules="rules">
                <el-form-item :label="$t('setting.versionItem')" prop="version">
                    <el-input class="p-w-200" disabled v-model="form.version">
                        <template #append>
                            <CopyButton class="w-16" :isIcon="false" :content="form.version" type="primary" />
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('setting.backupCopies')" prop="backupCopies">
                    <el-input class="p-w-200" type="number" v-model.number="form.backupCopies">
                        <template #append>
                            <el-button @click="onSave(formRef)" class="w-16">{{ $t('commons.button.save') }}</el-button>
                        </template>
                    </el-input>
                    <span class="input-help">{{ $t('setting.backupCopiesHelper') }}</span>
                </el-form-item>
            </el-form>
            <el-collapse v-if="notes && notes.length !== 0" v-model="currentVersion" :accordion="true">
                <div v-for="(item, index) in notes" :key="index">
                    <el-collapse-item :name="index">
                        <template #title>
                            <span class="version">{{ item.version }}</span>
                            <span v-if="!mobile" class="date">{{ item.createdAt }}</span>
                            <svg-icon class="icon" iconName="p-featureshitu"></svg-icon>
                            <span class="icon-span">{{ item.newCount }}</span>
                            <svg-icon class="icon" iconName="p-youhuawendang"></svg-icon>
                            <span class="icon-span">{{ item.optimizationCount }}</span>
                            <svg-icon class="icon" iconName="p-bug"></svg-icon>
                            <span class="icon-span">{{ item.fixCount }}</span>
                        </template>
                        <div class="panel-MdEditor">
                            <MdEditor v-model="item.content" previewOnly :theme="isDarkTheme ? 'dark' : 'light'" />
                        </div>
                    </el-collapse-item>
                </div>
            </el-collapse>
            <el-empty v-else>
                <template #description>
                    <span class="input-help">
                        {{ $t('setting.releaseHelper') }}
                        <el-link class="pageRoute" icon="Position" type="primary">
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                </template>
            </el-empty>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { getSettingInfo, listReleases, updateSetting } from '@/api/modules/setting';
import MdEditor from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import { ref } from 'vue';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { FormInstance } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';

const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const { isDarkTheme } = storeToRefs(globalStore);

const drawerVisible = ref(false);
const currentVersion = ref(0);
const notes = ref([]);
const loading = ref();
const formRef = ref();

const form = reactive({
    version: '',
    backupCopies: 0,
});
const rules = reactive({
    version: [Rules.requiredInput],
    backupCopies: [{ validator: checkBackupCopies, trigger: 'blur', required: true }],
});

const acceptParams = (): void => {
    search();
    loadInfo();
    drawerVisible.value = true;
};

const loadInfo = async () => {
    const res = await getSettingInfo();
    form.version = res.data.systemVersion;
    form.backupCopies = Number(res.data.upgradeBackupCopies) || 0;
};

function checkBackupCopies(rule: any, value: any, callback: any) {
    if (value === 0) {
        return callback();
    }
    if (value < 3) {
        return callback(new Error(i18n.global.t('setting.backupCopiesRule')));
    }
    callback();
}

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateSetting({ key: 'UpgradeBackupCopies', value: form.backupCopies + '' })
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

const search = async () => {
    loading.value = true;
    await listReleases()
        .then((res) => {
            notes.value = res.data || [];
            loading.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.version {
    margin-left: 10px;
    display: inline-block;
    width: 50px;
}
.date {
    margin-left: 20px;
    margin-right: 40px;
    display: inline-block;
    width: 100px;
}
.icon-span {
    display: inline-block;
    width: 10px;
}
.panel-MdEditor {
    :deep(.md-editor-preview) {
        font-size: 12px;
    }
}
:deep(.md-editor-dark) {
    background-color: var(--panel-main-bg-color-9);
}
:deep(.el-collapse-item__content) {
    padding: 0px;
}
.icon {
    display: inline-block;
    font-size: 7px;
    margin-left: 50px;
}
.pageRoute {
    font-size: 12px;
    margin-left: 5px;
    margin-top: -4px;
}
</style>
