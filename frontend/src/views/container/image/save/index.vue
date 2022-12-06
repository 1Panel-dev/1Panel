<template>
    <el-dialog v-model="saveVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.exportImage') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item label="Tag" :rules="Rules.requiredSelect" prop="tagName">
                <el-select filterable v-model="form.tagName">
                    <el-option
                        :disabled="item.indexOf(':<none>') !== -1"
                        v-for="item in form.tags"
                        :key="item"
                        :value="item"
                        :label="item"
                    />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.path')" :rules="Rules.requiredSelect" prop="path">
                <el-input clearable v-model="form.path">
                    <template #append>
                        <FileList @choose="loadSaveDir" :dir="true"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('container.fileName')" :rules="Rules.requiredInput" prop="name">
                <el-input v-model="form.name">
                    <template #append>.tar</template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="saveVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.export') }}
                </el-button>
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
import { imageSave } from '@/api/modules/container';
import { Container } from '@/api/interface/container';

const saveVisiable = ref(false);
const form = reactive({
    tags: [] as Array<string>,
    tagName: '',
    path: '',
    name: '',
});

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    tags: Array<string>;
}
const dialogData = ref<DialogProps>({
    repos: [] as Array<Container.RepoOptions>,
    tags: [] as Array<string>,
});

const acceptParams = async (params: DialogProps): Promise<void> => {
    saveVisiable.value = true;
    form.tags = params.tags;
    form.path = '';
    form.tagName = '';
    form.name = '';
    dialogData.value.repos = params.repos;
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            saveVisiable.value = false;
            await imageSave(form);
            emit('search');
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            emit('search');
        }
    });
};

const loadSaveDir = async (path: string) => {
    form.path = path;
};

defineExpose({
    acceptParams,
});
</script>
