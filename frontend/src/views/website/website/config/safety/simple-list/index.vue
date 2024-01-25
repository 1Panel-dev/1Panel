<template>
    <el-row>
        <el-col :xs="24" :sm="18" :md="10" :lg="10" :xl="10">
            <el-form-item prop="enable" :label="$t('website.enable')">
                <el-switch v-model="enableUpdate.enable" @change="updateEnable"></el-switch>
            </el-form-item>
            <el-form-item>
                <el-alert
                    type="info"
                    v-if="rule === 'ip_white'"
                    :title="$t('website.ipWhiteListHelper')"
                    :closable="false"
                />
            </el-form-item>
            <el-form-item :label="$t('website.ipValue')">
                <el-input type="textarea" :rows="3" v-model="ips" :placeholder="$t('website.wafInputHelper')" />
            </el-form-item>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column :label="$t('website.ipValue')" prop="ip"></el-table-column>
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
import { GetWafConfig, UpdateWafEnable, UpdateWafFile } from '@/api/modules/website';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { checkIpV4V6 } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';
import { MsgError } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
    rule: {
        type: String,
        default: 'ip_white',
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

const loading = ref(false);
const data = ref([]);
const req = ref<Website.WafReq>({
    websiteId: 0,
    key: '$ipWhiteAllow',
    rule: 'ip_white',
});
const fileUpdate = reactive({
    content: '',
    websiteId: 0,
    type: 'ip_white',
});
const enableUpdate = ref<Website.WafUpdate>({
    websiteId: 0,
    key: '$ipWhiteAllow',
    enable: false,
});
const ips = ref();

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
};

const removeIp = (index: number) => {
    const copyList = data.value.concat();
    copyList.splice(index, 1);
    let ipArray = [];
    copyList.forEach((d) => {
        ipArray.push(d.ip);
    });
    submit(ipArray);
};

const openCreate = () => {
    const ipArray = ips.value.split('\n');
    if (ipArray.length == 0) {
        return;
    }
    let newIpArray = [];
    ipArray.forEach((ip: string) => {
        const newIp = ip.replace(/(^\s*)|(\s*$)/g, '');
        if (newIp != '') {
            newIpArray.push(newIp);
        }
    });
    if (newIpArray.length == 0) {
        return;
    }
    if (req.value.rule.indexOf('ip') > -1) {
        for (const id in newIpArray) {
            if (checkIpV4V6(newIpArray[id])) {
                MsgError(i18n.global.t('commons.rule.ipErr', [ipArray[id]]));
                return;
            }
        }
    }
    data.value.forEach((d) => {
        newIpArray.push(d.ip);
    });
    if (newIpArray.length > 0) {
        submit(newIpArray);
    }
};

const submit = async (ipList: string[]) => {
    fileUpdate.content = JSON.stringify(ipList);
    loading.value = true;
    UpdateWafFile(fileUpdate)
        .then(() => {
            ips.value = '';
            get();
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateEnable = async (enable: boolean) => {
    enableUpdate.value.enable = enable;
    loading.value = true;
    try {
        await UpdateWafEnable(enableUpdate.value);
    } catch (error) {
        enableUpdate.value.enable = !enable;
    }
    loading.value = false;
};

onMounted(() => {
    req.value.websiteId = id.value;
    req.value.rule = rule.value;
    req.value.key = key.value;
    enableUpdate.value.websiteId = id.value;
    enableUpdate.value.key = key.value;
    fileUpdate.type = rule.value;
    fileUpdate.websiteId = id.value;
    get();
});
</script>
