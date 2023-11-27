<template>
    <div>
        <el-dialog v-model="open" :title="form.title" width="30%" :close-on-click-modal="false" @close="handleClose">
            <div v-loading="loading">
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
                    <el-button @click="open = false" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="onConfirm" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { onMounted, reactive, ref } from 'vue';

defineOptions({ name: 'OpDialog' });

const form = reactive({
    msgs: [],
    title: '',
    names: [],
    api: null as Function,
    params: {},
});
const loading = ref();
const open = ref();
const successMsg = ref('');

interface DialogProps {
    title: string;
    msg: string;
    names: Array<string>;

    api: Function;
    params: Object;
    successMsg: string;
}
const acceptParams = (props: DialogProps): void => {
    form.title = props.title;
    form.names = props.names;
    form.msgs = props.msg.split('\n');
    form.api = props.api;
    form.params = props.params;
    successMsg.value = props.successMsg;
    open.value = true;
};

const emit = defineEmits(['search', 'cancel', 'submit']);

const onConfirm = async () => {
    if (form.api === null) {
        emit('submit');
        open.value = false;
        return;
    }
    loading.value = true;
    await form
        .api(form.params)
        .then(() => {
            emit('cancel');
            emit('search');
            if (successMsg.value) {
                MsgSuccess(successMsg.value);
            } else {
                MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            }
            open.value = false;
            loading.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    emit('cancel');
    open.value = false;
};

onMounted(() => {});

defineExpose({
    acceptParams,
});
</script>
