<template>
    <div>
        <el-dialog v-model="open" :title="form.title" width="30%" :close-on-click-modal="false">
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
import { onMounted, reactive, ref } from 'vue';

const form = reactive({
    msgs: [],
    title: '',
    names: [],
    api: null as Function,
    params: {},
});
const loading = ref();
const open = ref();

interface DialogProps {
    title: string;
    msg: string;
    names: Array<string>;

    api: Function;
    params: Object;
}
const acceptParams = (props: DialogProps): void => {
    form.title = props.title;
    form.names = props.names;
    form.msgs = props.msg.split('\n');
    form.api = props.api;
    form.params = props.params;
    open.value = true;
};

const emit = defineEmits(['search', 'cancel']);

const onConfirm = async () => {
    if (form.api) {
        loading.value = true;
        await form
            .api(form.params)
            .then(() => {
                emit('cancel');
                emit('search');
                open.value = false;
                loading.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    }
};

onMounted(() => {});

defineExpose({
    acceptParams,
});
</script>
