<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('container.composeTemplate')"
        :resource="dialogData.title === 'create' ? '' : dialogData.rowData?.name"
        @close="handleClose"
        size="large"
    >
        <el-form
            v-loading="loading"
            label-position="top"
            ref="formRef"
            :model="dialogData.rowData"
            :rules="rules"
            label-width="80px"
        >
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input :disabled="dialogData.title === 'edit'" v-model.trim="dialogData.rowData!.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model="dialogData.rowData!.description"></el-input>
            </el-form-item>
            <el-form-item>
                <CodemirrorPro
                    placeholder="#Define or paste the content of your docker-compose file here"
                    v-model="dialogData.rowData!.content"
                    mode="yaml"
                    :heightDiff="400"
                ></CodemirrorPro>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Container } from '@/api/interface/container';
import { createComposeTemplate, updateComposeTemplate } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';

const loading = ref(false);

interface DialogProps {
    title: string;
    rowData?: Container.TemplateInfo;
    getTableList?: () => Promise<any>;
}
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
    name: [Rules.requiredInput, Rules.name],
    content: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await createComposeTemplate(dialogData.value.rowData!)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await updateComposeTemplate(dialogData.value.rowData!)
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
