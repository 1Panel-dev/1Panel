<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="40%"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.create')" :back="handleClose" />
        </template>

        <el-row>
            <el-col :span="22" :offset="1">
                <el-form
                    ref="fileForm"
                    label-position="top"
                    :model="addForm"
                    label-width="100px"
                    :rules="rules"
                    v-loading="loading"
                    @submit.enter.prevent
                >
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model="addForm.name" />
                    </el-form-item>
                    <el-form-item v-if="!addForm.isDir">
                        <el-checkbox v-model="addForm.isLink" :label="$t('file.link')"></el-checkbox>
                    </el-form-item>
                    <el-form-item :label="$t('file.linkType')" v-if="addForm.isLink" prop="linkType">
                        <el-radio-group v-model="addForm.isSymlink">
                            <el-radio :value="true">{{ $t('file.softLink') }}</el-radio>
                            <el-radio :value="false">{{ $t('file.hardLink') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item v-if="addForm.isLink" :label="$t('file.linkPath')" prop="linkPath">
                        <el-input v-model="addForm.linkPath">
                            <template #prepend>
                                <FileList @choose="getLinkPath"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-checkbox
                            v-if="addForm.isDir"
                            v-model="setRole"
                            :label="$t('file.editPermissions')"
                        ></el-checkbox>
                    </el-form-item>
                </el-form>
                <FileRole v-if="setRole" :mode="'0755'" @get-mode="getMode" :key="open.toString()"></FileRole>
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script setup lang="ts">
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ref, reactive, computed } from 'vue';
import { File } from '@/api/interface/file';
import { FormInstance, FormRules } from 'element-plus';
import { CreateFile } from '@/api/modules/files';
import i18n from '@/lang';
import FileRole from '@/components/file-role/index.vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import { MsgSuccess, MsgWarning } from '@/utils/message';

const fileForm = ref<FormInstance>();
let loading = ref(false);
let setRole = ref(false);

interface CreateProps {
    file: Object;
}
const propData = ref<CreateProps>({
    file: {},
});

let addForm = reactive({ path: '', name: '', isDir: false, mode: 0o755, isLink: false, isSymlink: true, linkPath: '' });
let open = ref(false);
const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
};

const rules = reactive<FormRules>({
    name: [Rules.requiredInput, Rules.linuxName],
    path: [Rules.requiredInput],
    isSymlink: [Rules.requiredInput],
    linkPath: [Rules.requiredInput],
});

const getMode = (val: number) => {
    addForm.mode = val;
};

let getPath = computed(() => {
    if (addForm.path.endsWith('/')) {
        return addForm.path + addForm.name.trim();
    } else {
        return addForm.path + '/' + addForm.name.trim();
    }
});

const getLinkPath = (path: string) => {
    addForm.linkPath = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        if (getPath.value.indexOf('.1panel_clash') > -1) {
            MsgWarning(i18n.global.t('file.clashDitNotSupport'));
            return;
        }

        let addItem = {};
        Object.assign(addItem, addForm);
        addItem['path'] = getPath.value;
        loading.value = true;
        if (!setRole.value) {
            addItem['mode'] = undefined;
        }
        addItem['name'] = addForm.name.trim();
        CreateFile(addItem as File.FileCreate)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const acceptParams = (create: File.FileCreate) => {
    propData.value.file = create;
    open.value = true;
    addForm.isDir = create.isDir;
    addForm.path = create.path;
    addForm.name = '';
    addForm.isLink = false;

    init();
};

const init = () => {
    setRole.value = false;
};

defineExpose({ acceptParams });
</script>
