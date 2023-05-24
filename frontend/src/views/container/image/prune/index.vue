<template>
    <el-dialog v-model="dialogVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.imagePrune') }}</span>
            </div>
        </template>
        <el-form ref="deleteForm" v-loading="loading">
            <el-form-item>
                <el-radio-group v-model="withTagAll">
                    <el-radio :label="false">{{ $t('container.imagePruneSome') }}</el-radio>
                    <el-radio :label="true">{{ $t('container.imagePruneAll') }}</el-radio>
                </el-radio-group>
                <span class="input-help">
                    {{ withTagAll ? $t('container.imagePruneAllHelper') : $t('container.imagePruneSomeHelper') }}
                </span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onClean" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { containerPrune } from '@/api/modules/container';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { computeSize } from '@/utils/util';
import { ref } from 'vue';

const dialogVisiable = ref(false);
const withTagAll = ref(false);
const loading = ref();

const acceptParams = (): void => {
    dialogVisiable.value = true;
    withTagAll.value = false;
};

const emit = defineEmits<{ (e: 'search'): void }>();

const onClean = async () => {
    loading.value = true;
    let params = {
        pruneType: 'image',
        withTagAll: withTagAll.value,
    };
    await containerPrune(params)
        .then((res) => {
            loading.value = false;
            dialogVisiable.value = false;
            MsgSuccess(
                i18n.global.t('container.cleanSuccessWithSpace', [
                    res.data.deletedNumber,
                    computeSize(res.data.spaceReclaimed),
                ]),
            );
            emit('search');
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
