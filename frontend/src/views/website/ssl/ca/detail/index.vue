<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('ssl.organizationDetail')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-radio-group v-model="curr">
                <el-radio-button value="detail">{{ $t('ssl.organizationDetail') }}</el-radio-button>
                <el-radio-button value="ssl">{{ $t('ssl.ssl') }}</el-radio-button>
                <el-radio-button value="key">{{ $t('ssl.key') }}</el-radio-button>
            </el-radio-group>
            <div v-if="curr === 'detail'" class="mt-5">
                <el-descriptions border :column="1">
                    <el-descriptions-item :label="$t('commons.table.name')">
                        {{ ca.name }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.commonName')">
                        {{ ca.commonName }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.brand')">
                        {{ ca.organization }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.department')">
                        {{ ca.organizationUint }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.country')">
                        {{ ca.country }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.province')">
                        {{ ca.province }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.city')">
                        {{ ca.city }}
                    </el-descriptions-item>
                </el-descriptions>
            </div>
            <div v-else-if="curr === 'ssl'" class="mt-5">
                <el-input v-model="ca.csr" :rows="15" type="textarea" id="textArea" />
                <div>
                    <br />
                    <CopyButton :content="ca.csr" />
                </div>
            </div>
            <div v-else class="mt-5">
                <el-input v-model="ca.privateKey" :rows="15" type="textarea" id="textArea" />
                <div>
                    <br />
                    <CopyButton :content="ca.privateKey" />
                </div>
            </div>
        </div>
    </el-drawer>
</template>
<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GetCA } from '@/api/modules/website';
import { ref } from 'vue';

const open = ref(false);
const id = ref(0);
const curr = ref('detail');
const ca = ref<any>({});
const loading = ref(false);

const handleClose = () => {
    open.value = false;
};

const acceptParams = (caID: number) => {
    ca.value = {};
    id.value = caID;
    curr.value = 'detail';
    get();
    open.value = true;
};

const get = async () => {
    const res = await GetCA(id.value);
    ca.value = res.data;
};

defineExpose({
    acceptParams,
});
</script>
