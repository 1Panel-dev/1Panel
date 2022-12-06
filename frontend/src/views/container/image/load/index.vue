<template>
    <el-dialog v-model="loadVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.importImage') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item :label="$t('container.path')" :rules="Rules.requiredSelect" prop="path">
                <el-input clearable v-model="form.path">
                    <template #append>
                        <FileList @choose="loadLoadDir" :dir="false"></FileList>
                    </template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="loadVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">{{ $t('container.import') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { imageLoad } from '@/api/modules/container';

const loadVisiable = ref(false);
const form = reactive({
    path: '',
});

const acceptParams = () => {
    loadVisiable.value = true;
    form.path = '';
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            loadVisiable.value = false;
            await imageLoad(form);
            emit('search');
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            emit('search');
        }
    });
};

const loadLoadDir = async (path: string) => {
    form.path = path;
};

defineExpose({
    acceptParams,
});
</script>
