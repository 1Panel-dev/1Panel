<template>
    <el-drawer v-model="open" :destroy-on-close="true" :close-on-click-modal="false" size="40%">
        <template #header>
            <DrawerHeader :header="title" :back="handleClose" />
        </template>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form
                    @submit.prevent
                    ref="fileForm"
                    label-position="top"
                    :model="addForm"
                    :rules="rules"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('file.path')" prop="newPath">
                        <el-input v-model="addForm.newPath">
                            <template #prepend><FileList @choose="getPath" :dir="true"></FileList></template>
                        </el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose(false)" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { MoveFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance, FormRules } from 'element-plus';
import { ref, reactive, computed } from 'vue';
import FileList from '@/components/file-list/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

interface MoveProps {
    oldPaths: Array<string>;
    type: string;
    path: string;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
let open = ref(false);
let type = ref('cut');

const title = computed(() => {
    if (type.value === 'cut') {
        return i18n.global.t('file.move');
    } else {
        return i18n.global.t('file.copy');
    }
});

const addForm = reactive({
    oldPaths: [] as string[],
    newPath: '',
    type: '',
});

const rules = reactive<FormRules>({
    newPath: [Rules.requiredInput],
});

const em = defineEmits(['close']);

const handleClose = (search: boolean) => {
    open.value = false;
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', search);
};

const getPath = (path: string) => {
    addForm.newPath = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        MoveFile(addForm)
            .then(() => {
                if (type.value === 'cut') {
                    MsgSuccess(i18n.global.t('file.moveSuccess'));
                } else {
                    MsgSuccess(i18n.global.t('file.copySuccess'));
                }
                handleClose(true);
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const acceptParams = (props: MoveProps) => {
    addForm.oldPaths = props.oldPaths;
    addForm.type = props.type;
    addForm.newPath = props.path;
    type.value = props.type;
    open.value = true;
};

defineExpose({ acceptParams });
</script>
