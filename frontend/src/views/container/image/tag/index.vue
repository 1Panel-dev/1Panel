<template>
    <DrawerPro v-model="drawerVisible" :header="$t('container.imageTag')" @close="handleClose" size="large">
        <el-form v-loading="loading" label-position="top" ref="formRef" :model="form" :rules="rules" label-width="80px">
            <el-form-item :label="$t('app.source')">
                <el-checkbox v-model="form.fromRepo">{{ $t('container.imageRepo') }}</el-checkbox>
            </el-form-item>
            <el-form-item
                v-if="form.fromRepo"
                :label="$t('container.repoName')"
                :rules="Rules.requiredSelect"
                prop="repo"
            >
                <el-select clearable filterable v-model="form.repo" @change="changeRepo">
                    <el-option v-for="item in repos" :key="item.id" :value="item.name" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.imageTag')" prop="tags">
                <el-input-tag ref="inputTagRef" @add-tag="handleAdd" v-model="form.tags">
                    <template #tag="{ value }">
                        <el-button @click="setInputValue(value)" size="small" link type="info">
                            {{ value }}
                        </el-button>
                    </template>
                </el-input-tag>
                <span class="input-help">{{ $t('container.imageTagHelper') }}</span>
            </el-form-item>
        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.save') }}
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
import { imageTag } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const inputTagRef = ref();

const drawerVisible = ref(false);
const repos = ref();
const form = reactive({
    imageID: '',
    fromRepo: false,
    repo: '',
    originName: '',

    tags: [],
});
const rules = reactive({
    tags: [{ validator: checkTags, trigger: 'blur', required: true }],
});
function checkTags(rule: any, value: any, callback: any) {
    if (value.length === 0) {
        return callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    }
    for (const item of value) {
        if (item === '' || typeof item === 'undefined' || item == null) {
            return callback(new Error(i18n.global.t('commons.rule.imageName')));
        } else {
            const reg = /^[a-zA-Z0-9]{1}[a-z:@A-Z0-9_/.-]{0,255}$/;
            if (!reg.test(item) && item !== '') {
                return callback(new Error(i18n.global.t('commons.rule.imageName')));
            } else {
                return callback();
            }
        }
    }
    callback();
}

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    imageID: string;
    tags: Array<string>;
}

const acceptParams = async (params: DialogProps): Promise<void> => {
    drawerVisible.value = true;
    form.imageID = params.imageID;
    form.originName = params.tags?.length !== 0 ? params.tags[0] : '';
    form.tags = params.tags || [];
    form.fromRepo = false;
    form.repo = '';
    repos.value = params.repos;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const handleAdd = (val: string) => {
    form.tags = form.tags?.filter((item) => item !== val);
    form.tags.push(val);
};
const setInputValue = async (text) => {
    await nextTick();
    const inputEl = inputTagRef.value?.$el?.querySelector('input');
    if (!inputEl) return;

    inputEl.value = text;
    inputEl.dispatchEvent(new Event('input', { bubbles: true }));
    inputEl.dispatchEvent(new Event('change', { bubbles: true }));
    inputEl.setSelectionRange(text.length, text.length);
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            sourceID: form.imageID,
            tags: form.tags,
        };
        loading.value = true;
        await imageTag(params)
            .then(async () => {
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const changeRepo = (val) => {
    if (val === 'Docker Hub') {
        return;
    }
    for (const item of repos.value) {
        if (item.name == val) {
            form.tags.push(item.downloadUrl + '/' + form.originName);
            return;
        }
    }
};

defineExpose({
    acceptParams,
});
</script>
