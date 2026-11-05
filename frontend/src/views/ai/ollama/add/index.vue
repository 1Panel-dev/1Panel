<template>
    <DrawerPro size="normal" v-model="drawerVisible" :header="$t('commons.button.add')" @close="handleClose">
        <el-alert type="info" :closable="false">
            <template #title>
                <i18n-t keypath="aiTools.model.ollama_doc" tag="span" class="ollama-hint-text">
                    <template #link>
                        <a
                            class="ollama-link"
                            href="https://ollama.com/search"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            ollama.com
                        </a>
                    </template>
                </i18n-t>
            </template>
        </el-alert>
        <el-form ref="formRef" label-position="top" class="mt-5" :model="form">
            <el-form-item :label="$t('commons.table.name')" :rules="Rules.requiredInput" prop="name">
                <el-input v-model.trim="form.name" placeholder="qwen3.6:35b" />
                <span class="input-help">{{ $t('aiTools.model.ollamaNameHelper') }}</span>
                <span class="input-help" v-if="form.name">
                    {{
                        $t('aiTools.model.create_helper', [
                            form.name.replaceAll('ollama run ', '').replaceAll('ollama pull ', ''),
                        ])
                    }}
                </span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button v-permission type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.add') }}
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
import { MsgSuccess } from '@/utils/message';
import { createAgentAccountModel, createOllamaModel } from '@/api/modules/ai';
import { newUUID } from '@/utils/id';
const props = defineProps<{ accountId?: number }>();
const drawerVisible = ref(false);
const form = reactive({
    name: '',
});

const acceptParams = async (): Promise<void> => {
    form.name = '';
    drawerVisible.value = true;
};
const emit = defineEmits(['search', 'log']);

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    let taskID = newUUID();
    formEl.validate(async (valid) => {
        if (!valid) return;
        let itemName = form.name.replaceAll('ollama run ', '').replaceAll('ollama pull ', '');
        const request = props.accountId
            ? createAgentAccountModel({
                  accountId: props.accountId,
                  model: { recordId: 0, id: itemName, name: itemName },
                  taskID,
              })
            : createOllamaModel(itemName, taskID);
        await request.then(() => {
            drawerVisible.value = false;
            emit('search');
            emit('log', { logFileExist: true, name: itemName, from: 'local', taskID: taskID });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.ollama-hint-text {
    font-size: 13px;
    font-weight: 400;
    line-height: 1.7;
}

.ollama-link {
    color: var(--el-color-primary);
    font: inherit;
    vertical-align: baseline;
    text-decoration: none;
}

.ollama-link:hover {
    text-decoration: underline;
    text-underline-offset: 2px;
}
</style>
