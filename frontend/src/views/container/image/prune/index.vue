<template>
    <el-dialog v-model="dialogVisible" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.imagePrune') }}</span>
            </div>
        </template>
        <el-form ref="deleteForm" v-loading="loading">
            <el-form-item>
                <el-radio-group v-model="withTagAll">
                    <el-radio :value="false">{{ $t('container.imagePruneSome') }}</el-radio>
                    <el-radio :value="true">{{ $t('container.imagePruneAll') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <span v-if="withTagAll">
                {{ unUsedList.length !== 0 ? $t('container.imagePruneAllHelper') : $t('container.imagePruneAllEmpty') }}
            </span>
            <span v-else>
                {{
                    unTagList.length !== 0 ? $t('container.imagePruneSomeHelper') : $t('container.imagePruneSomeEmpty')
                }}
            </span>
            <div v-if="!withTagAll">
                <ul v-for="(item, index) in unTagList" :key="index">
                    <li v-if="item.tags && item.tags[0]">
                        {{ item.tags[0] }}
                    </li>
                    <li v-else>
                        {{ item.id.replaceAll('sha256:', '').substring(0, 12) }}
                    </li>
                </ul>
            </div>
            <div v-else>
                <ul v-for="(item, index) in unUsedList" :key="index">
                    <li v-if="item.tags && item.tags[0]">{{ item.tags.join(', ') }}</li>
                    <li v-else>{{ item.id.replaceAll('sha256:', '').substring(0, 12) }}</li>
                </ul>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="buttonDisable() || loading" @click="onClean">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { containerPrune, listAllImage } from '@/api/modules/container';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { computeSize } from '@/utils/util';
import { ref } from 'vue';

const dialogVisible = ref(false);
const withTagAll = ref(false);
const loading = ref();
const unTagList = ref();
const unUsedList = ref();

const acceptParams = async (): Promise<void> => {
    const res = await listAllImage();
    let list = res.data || [];
    unTagList.value = [];
    unUsedList.value = [];
    for (const item of list) {
        if (
            !item.tags ||
            item.tags.length === 0 ||
            (item.tags.length === 1 && item.tags[0].indexOf('<none>') !== -1 && !item.isUsed)
        ) {
            unTagList.value.push(item);
        }
        if (!item.isUsed) {
            unUsedList.value.push(item);
        }
    }
    dialogVisible.value = true;
    withTagAll.value = false;
};

const emit = defineEmits<{ (e: 'search'): void }>();

const buttonDisable = () => {
    return withTagAll.value ? unUsedList.value.length === 0 : unTagList.value.length === 0;
};

const onClean = async () => {
    loading.value = true;
    let params = {
        pruneType: 'image',
        withTagAll: withTagAll.value,
    };
    await containerPrune(params)
        .then((res) => {
            loading.value = false;
            dialogVisible.value = false;
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
