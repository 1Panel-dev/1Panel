<template>
    <el-drawer
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="40%"
    >
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
                    <div v-if="changeName">
                        <el-form-item :label="$t('commons.table.name')" prop="name">
                            <el-input v-model="addForm.name" :disabled="addForm.cover"></el-input>
                        </el-form-item>
                        <el-radio-group v-model="addForm.cover" @change="changeType">
                            <el-radio :value="true" size="large">{{ $t('file.replace') }}</el-radio>
                            <el-radio :value="false" size="large">{{ $t('file.rename') }}</el-radio>
                        </el-radio-group>
                    </div>
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
import { CheckFile, MoveFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance, FormRules } from 'element-plus';
import { ref, reactive, computed } from 'vue';
import FileList from '@/components/file-list/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { getDateStr } from '@/utils/util';

interface MoveProps {
    oldPaths: Array<string>;
    type: string;
    path: string;
    name: string;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
const open = ref(false);
const type = ref('cut');
const changeName = ref(false);
const oldName = ref('');

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
    name: '',
    cover: false,
});

const rules = reactive<FormRules>({
    newPath: [Rules.requiredInput],
    name: [Rules.requiredInput],
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

const changeType = () => {
    if (addForm.cover) {
        addForm.name = oldName.value;
    } else {
        addForm.name = oldName.value + '-' + getDateStr();
    }
};

const mvFile = () => {
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
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        mvFile();
    });
};

const acceptParams = async (props: MoveProps) => {
    changeName.value = false;
    addForm.oldPaths = props.oldPaths;
    addForm.type = props.type;
    addForm.newPath = props.path;
    addForm.name = '';
    type.value = props.type;
    if (props.name && props.name != '') {
        oldName.value = props.name;
        const res = await CheckFile(props.path + '/' + props.name);
        if (res.data) {
            changeName.value = true;
            addForm.cover = false;
            addForm.name = props.name + '-' + getDateStr();
            open.value = true;
        } else {
            mvFile();
        }
    } else {
        mvFile();
    }
};

defineExpose({ acceptParams });
</script>
