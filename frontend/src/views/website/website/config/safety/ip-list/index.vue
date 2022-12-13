<template>
    <el-row>
        <el-col :span="10" :offset="2">
            <el-form-item prop="enable" :label="$t('website.enable')">
                <el-switch v-model="enableUpdate.enable" @change="updateEnable"></el-switch>
            </el-form-item>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column label="IP" prop="ip">
                    <template #default="{ row }">
                        <fu-read-write-switch :data="row.ip" v-model="row.edit" write-trigger="onDblclick">
                            <el-form-item :error="row.error">
                                <el-input v-model="row.ip" @blur="row.edit = false" @input="checkIpRule(row)" />
                            </el-form-item>
                        </fu-read-write-switch>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')">
                    <template #default="{ $index }">
                        <el-button link type="primary" @click="removeIp($index)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </ComplexTable>
            <br />
            <el-alert :title="$t('website.mustSave')" type="info" :closable="false"></el-alert>
            <br />
            <el-button type="primary" :loading="loading" @click="submit">
                {{ $t('commons.button.save') }}
            </el-button>
        </el-col>
    </el-row>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { GetWafConfig, UpdateWafEnable } from '@/api/modules/website';
import { computed, onMounted, reactive, ref } from 'vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { SaveFileContent } from '@/api/modules/files';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import { checkIp } from '@/utils/util';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
    rule: {
        type: String,
        default: 'ipWhitelist',
    },
    paramKey: {
        type: String,
        default: '$ipWhiteAllow',
    },
});
const id = computed(() => {
    return props.id;
});
const rule = computed(() => {
    return props.rule;
});
const key = computed(() => {
    return props.paramKey;
});

let loading = ref(false);
let data = ref([]);
let req = ref<Website.WafReq>({
    websiteId: 0,
    key: '$ipWhiteAllow',
    rule: 'ipWhitelist',
});
let fileUpdate = reactive({
    path: '',
    content: '',
});
let enableUpdate = ref<Website.WafUpdate>({
    websiteId: 0,
    key: '$ipWhiteAllow',
    enable: false,
});

const get = async () => {
    data.value = [];
    loading.value = true;
    const res = await GetWafConfig(req.value);
    loading.value = false;

    if (res.data.content != '') {
        const ipList = JSON.parse(res.data.content);
        ipList.forEach((value) => {
            data.value.push({
                ip: value,
                eidt: false,
                error: '',
            });
        });
    }
    enableUpdate.value.enable = res.data.enable;
    fileUpdate.path = res.data.filePath;
};

const removeIp = (index: number) => {
    data.value.splice(index, 1);
};

const openCreate = () => {
    data.value.unshift({ ip: '', edit: true, error: '' });
};

const submit = async () => {
    let canCommit = true;
    for (const row of data.value) {
        if (row.ip != '' && row.error != '') {
            row.edit = true;
            canCommit = false;
        }
    }
    if (!canCommit) {
        return;
    }
    let ipArray = [];
    data.value.forEach((row) => {
        ipArray.push(row.ip);
    });

    fileUpdate.content = JSON.stringify(ipArray);
    loading.value = true;
    SaveFileContent(fileUpdate)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateEnable = async (enable: boolean) => {
    enableUpdate.value.enable = enable;
    loading.value = true;
    await UpdateWafEnable(enableUpdate.value);
    loading.value = false;
};

const checkIpRule = (row: any) => {
    if (checkIp(row.ip)) {
        row.error = i18n.global.t('commons.rule.ip');
    } else {
        row.error = '';
    }
};

onMounted(() => {
    req.value.websiteId = id.value;
    req.value.rule = rule.value;
    req.value.key = key.value;
    enableUpdate.value.websiteId = id.value;
    enableUpdate.value.key = key.value;
    get();
});
</script>
