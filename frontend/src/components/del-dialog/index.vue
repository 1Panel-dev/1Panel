<template>
    <div>
        <component
            :is="drawer ? DrawerPro : DialogPro"
            v-model="open"
            v-bind="drawer ? { header: form.title, size: 'large' } : { title: form.title }"
            @close="handleClose"
        >
            <div v-loading="loading" class="del-dialog-body">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert class="mt-2" :show-icon="true" type="warning" :closable="false">
                            <div v-for="(item, index) in form.msgs" :key="index">
                                <div style="line-height: 20px; word-wrap: break-word">
                                    <span>{{ item }}</span>
                                </div>
                            </div>
                        </el-alert>
                        <slot name="content"></slot>
                        <ul v-for="(item, index) in form.names" :key="index">
                            <div style="word-wrap: break-word">
                                <li>{{ item }}</li>
                            </div>
                        </ul>
                    </el-col>
                </el-row>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="onConfirm" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </component>
    </div>
</template>

<script setup lang="ts">
import DialogPro from '@/components/dialog-pro/index.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
withDefaults(defineProps<{ drawer?: boolean }>(), { drawer: false });
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { reactive, ref } from 'vue';

defineOptions({ name: 'OpDialog' });

const form = reactive({
    msgs: [],
    title: '',
    names: [],
    api: null as Function,
    params: {},
});
const loading = ref(false);
const open = ref(false);
const successMsg = ref('');
const noMsg = ref(false);

interface DialogProps {
    title: string;
    msg: string;
    names: Array<string>;

    api: Function;
    params: object;
    successMsg?: string;
    noMsg?: boolean;
}
const acceptParams = (props: DialogProps): void => {
    form.title = props.title;
    form.names = props.names;
    form.msgs = props.msg.split('\n');
    form.api = props.api;
    form.params = props.params;
    successMsg.value = props.successMsg;
    noMsg.value = props.noMsg ?? false;
    open.value = true;
};

const emit = defineEmits(['search', 'cancel', 'submit']);

const onConfirm = async () => {
    if (loading.value) return;
    if (form.api === null) {
        emit('submit');
        open.value = false;
        return;
    }
    loading.value = true;
    try {
        await form.api(form.params);
        emit('search');
        if (!noMsg.value) {
            MsgSuccess(successMsg.value ?? i18n.global.t('commons.msg.deleteSuccess'));
        }
        open.value = false;
    } catch {
        // API errors are reported by the request interceptor; keep the dialog open for retry.
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    emit('cancel');
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.del-dialog-body {
    max-height: 60vh;
    overflow-y: auto;
}
</style>
