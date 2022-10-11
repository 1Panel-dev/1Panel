<template>
    <el-dialog v-model="buildVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.importImage') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item label="Dockerfile" :rules="Rules.requiredSelect" prop="from">
                <el-radio-group v-model="form.from">
                    <el-radio label="edit">{{ $t('container.edit') }}</el-radio>
                    <el-radio label="path">{{ $t('container.pathSelect') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.from === 'edit'" :rules="Rules.requiredInput">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 10 }" v-model="form.dockerfile" />
            </el-form-item>
            <el-form-item v-else :rules="Rules.requiredInput">
                <el-input clearable v-model="form.dockerfile">
                    <template #append>
                        <FileList @choose="loadBuildDir" :dir="true"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('container.tag')">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" v-model="form.tag" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onSubmit(formRef)">{{ $t('container.import') }}</el-button>
                <el-button @click="buildVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { imageBuild } from '@/api/modules/container';

const buildVisiable = ref(false);
const form = reactive({
    from: 'path',
    dockerfile: '',
    tag: '',
});
const acceptParams = async () => {
    buildVisiable.value = true;
    form.from = 'path';
    form.dockerfile = '';
    form.tag = '';
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            buildVisiable.value = false;
            await imageBuild(form);
            emit('search');
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            emit('search');
        }
    });
};

const loadBuildDir = async (path: string) => {
    form.dockerfile = path;
};

defineExpose({
    acceptParams,
});
</script>
