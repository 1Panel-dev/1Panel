<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader
                :header="title"
                :hideResource="dialogData.title === 'add'"
                :resource="dialogData.rowData?.name"
                :back="handleClose"
            />
        </template>
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input
                            :disabled="dialogData.title === 'edit'"
                            clearable
                            v-model.trim="dialogData.rowData!.name"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('toolbox.clam.scanDir')" prop="path">
                        <el-input v-model="dialogData.rowData!.path">
                            <template #prepend>
                                <FileList @choose="loadDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item :label="$t('toolbox.clam.infectedStrategy')" prop="infectedStrategy">
                        <el-radio-group v-model="dialogData.rowData!.infectedStrategy">
                            <el-radio value="none">{{ $t('toolbox.clam.none') }}</el-radio>
                            <el-radio value="remove">{{ $t('toolbox.clam.remove') }}</el-radio>
                            <el-radio value="move">{{ $t('toolbox.clam.move') }}</el-radio>
                            <el-radio value="copy">{{ $t('toolbox.clam.copy') }}</el-radio>
                        </el-radio-group>
                        <span class="input-help">
                            {{ $t('toolbox.clam.' + dialogData.rowData!.infectedStrategy + 'Helper') }}
                        </span>
                    </el-form-item>
                    <el-form-item v-if="hasInfectedDir()" :label="$t('toolbox.clam.infectedDir')" prop="infectedDir">
                        <el-input v-model="dialogData.rowData!.infectedDir">
                            <template #prepend>
                                <FileList @choose="loadInfectedDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.description')" prop="description">
                        <el-input type="textarea" :rows="3" clearable v-model="dialogData.rowData!.description" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Toolbox } from '@/api/interface/toolbox';
import { createClam, updateClam } from '@/api/modules/toolbox';

interface DialogProps {
    title: string;
    rowData?: Toolbox.ClamInfo;
    getTableList?: () => Promise<any>;
}
const loading = ref();
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    name: [Rules.simpleName],
    path: [Rules.requiredInput, Rules.noSpace],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const hasInfectedDir = () => {
    return (
        dialogData.value.rowData!.infectedStrategy === 'move' || dialogData.value.rowData!.infectedStrategy === 'copy'
    );
};
const loadDir = async (path: string) => {
    dialogData.value.rowData!.path = path;
};
const loadInfectedDir = async (path: string) => {
    dialogData.value.rowData!.infectedDir = path;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (dialogData.value.title === 'edit') {
            await updateClam(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    drawerVisible.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });

            return;
        }

        await createClam(dialogData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
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
