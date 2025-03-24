<template>
    <div v-loading="loading">
        <LayoutContent backName="Clam">
            <template #app>
                <ClamStatus v-model:loading="loading" />
            </template>
            <template #leftToolBar>
                <el-button type="primary" :plain="activeName !== 'clamd'" @click="search('clamd')">
                    {{ $t('toolbox.clam.clamConf') }}
                </el-button>
                <el-button type="primary" :plain="activeName !== 'freshclam'" @click="search('freshclam')">
                    {{ $t('toolbox.clam.freshClam') }}
                </el-button>
                <el-button type="primary" :plain="activeName !== 'clamd-log'" @click="search('clamd-log')">
                    {{ $t('toolbox.clam.clamLog') }}
                </el-button>
                <el-button type="primary" :plain="activeName !== 'freshclam-log'" @click="search('freshclam-log')">
                    {{ $t('toolbox.clam.freshClamLog') }}
                </el-button>
            </template>

            <template #main>
                <div>
                    <el-select v-if="!canUpdate()" style="width: 20%" @change="search()" v-model.number="tail">
                        <template #prefix>{{ $t('toolbox.clam.tail') }}</template>
                        <el-option :value="0" :label="$t('commons.table.all')" />
                        <el-option :value="10" :label="10" />
                        <el-option :value="100" :label="100" />
                        <el-option :value="200" :label="200" />
                        <el-option :value="500" :label="500" />
                        <el-option :value="1000" :label="1000" />
                    </el-select>
                    <CodemirrorPro
                        :heightDiff="400"
                        class="mt-5"
                        v-model="content"
                        :disabled="!canUpdate()"
                        :placeholder="$t('commons.msg.noneData')"
                    ></CodemirrorPro>
                    <el-button type="primary" class="mt-5" v-if="canUpdate()" @click="onSave">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmRef" @confirm="onSubmit"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import ClamStatus from '@/views/toolbox/clam/status/index.vue';
import { searchClamFile, updateClamFile } from '@/api/modules/toolbox';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const activeName = ref('clamd');
const tail = ref(200);
const content = ref();
const confirmRef = ref();

const canUpdate = () => {
    return activeName.value.indexOf('-log') === -1;
};

const search = async (itemName?: string) => {
    if (itemName) {
        tail.value = itemName.indexOf('-log') === -1 ? 0 : 200;
        activeName.value = itemName;
    }
    loading.value = true;
    await searchClamFile(activeName.value, tail.value + '')
        .then((res) => {
            loading.value = false;
            content.value = res.data;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSave = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmRef.value!.acceptParams(params);
};

const onSubmit = async () => {
    loading.value = true;
    await updateClamFile(activeName.value, content.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search(activeName.value);
});
</script>
