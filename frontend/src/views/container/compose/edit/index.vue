<template>
    <DrawerPro
        v-model="composeVisible"
        :header="$t('commons.button.edit')"
        @close="handleClose"
        :resource="name"
        size="large"
    >
        <div v-loading="loading">
            <el-form ref="formRef" @submit.prevent label-position="top">
                <el-form-item>
                    <CodemirrorPro
                        v-model="content"
                        mode="yaml"
                        :heightDiff="225"
                        placeholder="#Define or paste the content of your docker-compose file here"
                    ></CodemirrorPro>
                </el-form-item>
                <div v-if="createdBy === '1Panel'">
                    <el-form-item :label="$t('container.env')" prop="environmentStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="environmentStr"
                        />
                    </el-form-item>
                    <span class="input-help whitespace-break-spaces">
                        {{ $t('container.editComposeHelper') }}
                    </span>
                    <CodemirrorPro
                        v-model="envFileContent"
                        :height="45"
                        :minHeight="45"
                        disabled
                        mode="yaml"
                    ></CodemirrorPro>
                </div>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="composeVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmitEdit()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { composeUpdate } from '@/api/modules/container';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const composeVisible = ref(false);
const path = ref();
const content = ref();
const name = ref();
const environmentStr = ref();
const environmentEnv = ref();
const createdBy = ref();
const envFileContent = ref(`env_file:\n  - 1panel.env`);

const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmitEdit = async () => {
    const param = {
        name: name.value,
        path: path.value,
        content: content.value,
        env: environmentStr.value,
        createdBy: createdBy.value,
    };
    if (environmentStr.value != undefined) {
        param.env = environmentStr.value.split('\n');
        emit('search');
    }
    loading.value = true;
    await composeUpdate(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            composeVisible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

interface DialogProps {
    name: string;
    path: string;
    content: string;
    env: Array<string>;
    envStr: string;
    createdBy: string;
}
const acceptParams = (props: DialogProps): void => {
    composeVisible.value = true;
    path.value = props.path;
    name.value = props.name;
    content.value = props.content;
    createdBy.value = props.createdBy;
    environmentEnv.value = props.env || [];
    environmentStr.value = environmentEnv.value.join('\n');
};
const handleClose = () => {
    composeVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
