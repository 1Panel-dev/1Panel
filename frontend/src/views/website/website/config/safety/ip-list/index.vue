<template>
    <el-row>
        <el-col :span="10" :offset="2">
            <el-form-item prop="enable" :label="$t('website.enable')">
                <el-switch v-model="enableUpdate.enable" @change="updateEnable"></el-switch>
            </el-form-item>
            <el-form-item :label="$t('website.ipValue')">
                <el-input
                    type="textarea"
                    :autosize="{ minRows: 4, maxRows: 8 }"
                    v-model="ips"
                    :placeholder="$t('website.wafInputHelper')"
                />
            </el-form-item>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column label="IP" prop="ip"></el-table-column>
                <el-table-column :label="$t('commons.table.operate')">
                    <template #default="{ $index }">
                        <el-button link type="primary" @click="removeIp($index)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </ComplexTable>
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
        default: 'ipWhiteList',
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
    rule: 'ipWhiteList',
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
let ips = ref();

const get = async () => {
    data.value = [];
    loading.value = true;
    const res = await GetWafConfig(req.value);
    loading.value = false;

    if (res.data.content != '') {
        const ipList = JSON.parse(res.data.content);
        ipList.forEach((ip: string) => {
            data.value.push({
                ip: ip,
            });
        });
    }
    enableUpdate.value.enable = res.data.enable;
    fileUpdate.path = res.data.filePath;
};

const removeIp = (index: number) => {
    data.value.splice(index, 1);
    let ipArray = [];
    data.value.forEach((d) => {
        ipArray.push(d.ip);
    });
    submit(ipArray);
};

const openCreate = () => {
    console.log(ips.value);
    const ipArray = ips.value.split('\n');
    if (ipArray.length == 0) {
        return;
    }
    for (const id in ipArray) {
        if (checkIp(ipArray[id])) {
            ElMessage.error(i18n.global.t('commons.rule.ipErr', [ipArray[id]]));
            return;
        }
    }

    data.value.forEach((d) => {
        ipArray.push(d.ip);
    });
    submit(ipArray);
};

const submit = async (ipList: string[]) => {
    fileUpdate.content = JSON.stringify(ipList);
    loading.value = true;
    SaveFileContent(fileUpdate)
        .then(() => {
            ips.value = '';
            get();
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

onMounted(() => {
    req.value.websiteId = id.value;
    req.value.rule = rule.value;
    req.value.key = key.value;
    enableUpdate.value.websiteId = id.value;
    enableUpdate.value.key = key.value;
    get();
});
</script>
