<template>
    <el-dialog
        v-model="open"
        :title="title"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :before-close="handleClose"
        width="30%"
        @open="onOpen"
    >
        <el-form
            ref="fileForm"
            label-position="left"
            :model="addForm"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('file.path')" prop="newPath">
                <el-input v-model="addForm.newPath" disabled>
                    <template #append><FileList @choose="getPath" :dir="true"></FileList></template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { MoveFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { toRefs, ref, reactive, PropType, computed } from 'vue';
import FileList from '@/components/file-list/index.vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    oldPaths: {
        type: Array as PropType<string[]>,
        default: () => {
            return [];
        },
    },
    type: {
        type: String,
        default: '',
    },
});

const { open } = toRefs(props);
const fileForm = ref<FormInstance>();
const loading = ref(false);

const title = computed(() => {
    if (props.type === 'cut') {
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

const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
};

const getPath = (path: string) => {
    console.log(path);
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
                ElMessage.success(i18n.global.t('file.moveStart'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const onOpen = () => {
    addForm.oldPaths = props.oldPaths;
    addForm.type = props.type;
};
</script>
