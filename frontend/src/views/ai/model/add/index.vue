<template>
    <DrawerPro v-model="drawerVisible" :header="$t('aiTools.model.create')" @close="handleClose">
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-alert type="info" :closable="false">
                    <template #title>
                        <span class="flx-align-center">
                            {{ $t('aiTools.model.ollama_doc') }}
                            <el-button link class="ml-5" icon="Position" @click="goSearch()" type="primary">
                                {{ $t('firewall.quickJump') }}
                            </el-button>
                        </span>
                    </template>
                </el-alert>
                <el-form ref="formRef" label-position="top" class="mt-5" :model="form">
                    <el-form-item :label="$t('commons.table.name')" :rules="Rules.requiredInput" prop="name">
                        <el-input v-model.trim="form.name" />
                        <span class="input-help" v-if="form.name">
                            {{
                                $t('aiTools.model.create_helper', [
                                    form.name.replaceAll('ollama run ', '').replaceAll('ollama pull ', ''),
                                ])
                            }}
                        </span>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
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
import { createOllamaModel } from '@/api/modules/ai';
import { newUUID } from '@/utils/util';

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
        await createOllamaModel(itemName, taskID).then(() => {
            drawerVisible.value = false;
            emit('search');
            emit('log', { logFileExist: true, name: itemName, from: 'local', taskID: taskID });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        });
    });
};

const goSearch = () => {
    window.open('https://ollama.com/search', '_blank', 'noopener,noreferrer');
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
