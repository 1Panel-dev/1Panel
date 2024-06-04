<template>
    <el-drawer
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :before-close="handleClose"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="title" :back="handleClose" />
        </template>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form
                    ref="fileForm"
                    label-position="top"
                    :model="form"
                    label-width="100px"
                    :rules="rules"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('file.compressType')" prop="type">
                        <el-select v-model="form.type">
                            <el-option v-for="item in options" :key="item" :label="item" :value="item" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model="form.name">
                            <template #append>{{ extension }}</template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('file.compressDst')" prop="dst">
                        <el-input v-model="form.dst">
                            <template #prepend>
                                <FileList :path="form.dst" @choose="getLinkPath" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('setting.compressPassword')" prop="secret" v-if="form.type === 'tar.gz'">
                        <el-input v-model="form.secret"></el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-checkbox v-model="form.replace" :label="$t('file.replace')"></el-checkbox>
                    </el-form-item>
                </el-form>
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
import i18n from '@/lang';
import { computed, reactive, ref } from 'vue';
import { File } from '@/api/interface/file';
import { FormInstance, FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { CompressExtension, CompressType } from '@/enums/files';
import { CompressFile } from '@/api/modules/files';
import FileList from '@/components/file-list/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

interface CompressProps {
    files: Array<any>;
    dst: string;
    name: string;
    operate: string;
}

const rules = reactive<FormRules>({
    type: [Rules.requiredSelect],
    dst: [Rules.requiredInput],
    name: [Rules.requiredInput],
});

const fileForm = ref<FormInstance>();
const loading = ref(false);
const form = ref<File.FileCompress>({ files: [], type: 'zip', dst: '', name: '', replace: false, secret: '' });
const options = ref<string[]>([]);
const open = ref(false);
const title = ref('');
const operate = ref('compress');

const em = defineEmits(['close']);

const extension = computed(() => {
    return CompressExtension[form.value.type];
});

const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
    open.value = false;
};

const getLinkPath = (path: string) => {
    form.value.dst = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        let addItem = {};
        Object.assign(addItem, form.value);
        addItem['name'] = form.value.name + extension.value;
        loading.value = true;
        CompressFile(addItem as File.FileCompress)
            .then(() => {
                MsgSuccess(i18n.global.t('file.compressSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const acceptParams = (props: CompressProps) => {
    form.value.files = props.files;
    form.value.dst = props.dst;
    form.value.name = props.name;

    operate.value = props.operate;
    options.value = [];
    for (const t in CompressType) {
        options.value.push(CompressType[t]);
    }
    open.value = true;

    title.value = i18n.global.t('file.' + props.operate);
};

defineExpose({ acceptParams });
</script>
