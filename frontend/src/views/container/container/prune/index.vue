<template>
    <DialogPro v-model="open" :title="$t('container.containerPrune')" size="small">
        <div>
            <ul class="help-ul">
                <li class="lineClass" style="color: red">{{ $t('container.containerPruneHelper1') }}</li>
                <li class="lineClass">{{ $t('container.containerPruneHelper2') }}</li>
                <li class="lineClass">{{ $t('container.containerPruneHelper3') }}</li>
            </ul>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="open = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onClean()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { containerPrune } from '@/api/modules/container';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';
import { computeSize } from '@/utils/util';

const loading = ref(false);
const open = ref<boolean>(false);

const emit = defineEmits<{ (e: 'search'): void }>();

const onClean = async () => {
    loading.value = true;
    let params = {
        pruneType: 'container',
        withTagAll: false,
    };
    await containerPrune(params)
        .then((res) => {
            loading.value = false;
            MsgSuccess(
                i18n.global.t('container.cleanSuccessWithSpace', [
                    res.data.deletedNumber,
                    computeSize(res.data.spaceReclaimed),
                ]),
            );
            open.value = false;
            emit('search');
        })
        .catch(() => {
            loading.value = false;
        });
};

const acceptParams = (): void => {
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.lineClass {
    line-height: 30px;
}
</style>
