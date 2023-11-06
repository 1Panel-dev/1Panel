<template>
    <el-row>
        <el-col :xs="24" :sm="18" :md="14" :lg="14" :xl="14">
            <el-form>
                <el-form-item prop="enable" :label="$t('website.enable')" v-if="rule != 'user_agent'">
                    <el-switch v-model="enableUpdate.enable" @change="updateEnable"></el-switch>
                </el-form-item>
                <el-form-item :label="$t('website.data')">
                    <el-row :gutter="10" style="width: 100%">
                        <el-col :span="12">
                            <el-input
                                type="text"
                                v-model="add.value"
                                label="value"
                                :placeholder="$t('website.wafValueHelper')"
                            />
                        </el-col>
                        <el-col :span="12">
                            <el-input
                                type="text"
                                v-model="add.remark"
                                label="remark"
                                :placeholder="$t('website.wafRemarkHelper')"
                            />
                        </el-col>
                    </el-row>
                </el-form-item>
            </el-form>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column :label="$t('website.value')" prop="value"></el-table-column>
                <el-table-column :label="$t('website.remark')" prop="remark"></el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="100px">
                    <template #default="{ $index }">
                        <el-button link type="primary" @click="remove($index)">
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
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
    rule: {
        type: String,
        default: 'url',
    },
    paramKey: {
        type: String,
        default: 'url',
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
    key: '',
    rule: 'url',
});
const fileUpdate = reactive({
    content: '',
    websiteId: 0,
    type: 'url',
});
const enableUpdate = ref<Website.WafUpdate>({
    websiteId: 0,
    key: '$UrlDeny',
    enable: false,
});
let add = ref({
    value: '',
    remark: '',
    enbale: 1,
});
let contentArray = ref([]);

const get = async () => {
    data.value = [];
    loading.value = true;
    const res = await GetWafConfig(req.value);
    loading.value = false;
    enableUpdate.value.enable = res.data.enable;
    if (res.data.content != '') {
        contentArray.value = JSON.parse(res.data.content);
        contentArray.value.forEach((value) => {
            if (value != '') {
                data.value.push({
                    value: value[0],
                    remark: value[1],
                    enable: value[2],
                });
            }
        });
    }
};

const remove = (index: number) => {
    contentArray.value.splice(index, 1);
    submit([]);
};

const openCreate = () => {
    if (add.value.value == '') {
        return;
    }
    let newArray = [];
    newArray[0] = add.value.value;
    newArray[1] = add.value.remark;
    newArray[2] = add.value.enbale;

    data.value.push(newArray);
    submit(newArray);
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

const submit = async (addArray: string[]) => {
    if (addArray.length > 0) {
        contentArray.value.push(addArray);
    }

    fileUpdate.content = JSON.stringify(contentArray.value);
    loading.value = true;
    UpdateWafFile(fileUpdate)
        .then(() => {
            add.value = {
                value: '',
                remark: '',
                enbale: 1,
            };
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            get();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    req.value.websiteId = id.value;
    req.value.rule = rule.value;
    req.value.key = key.value;
    enableUpdate.value.key = key.value;
    enableUpdate.value.websiteId = id.value;
    fileUpdate.websiteId = id.value;
    fileUpdate.type = rule.value;
    get();
});
</script>
