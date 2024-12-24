<template>
    <DialogPro v-model="open" :title="$t('commons.button.edit')" size="small">
        <div v-if="isFromApp" class="leading-6">
            <div>
                <span>{{ $t('container.updateHelper1') }}</span>
            </div>
            <br />
            <div>
                <span>{{ $t('container.updateHelper2') }}</span>
            </div>
            <div>
                <span>{{ $t('container.updateHelper3') }}</span>
            </div>
            <br />
        </div>
        <div>
            <span>{{ $t('container.updateHelper4') }}</span>
        </div>
        <template #footer>
            <el-button :disabled="loading" @click="open = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="onSubmit()">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const loading = ref();
const open = ref(false);
const isFromApp = ref();

interface DialogProps {
    isFromApp: boolean;
}

const acceptParams = (props: DialogProps): void => {
    isFromApp.value = props.isFromApp;
    open.value = true;
};
const emit = defineEmits(['submit']);

const onSubmit = async () => {
    emit('submit');
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>
