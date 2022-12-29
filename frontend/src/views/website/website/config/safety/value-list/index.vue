<template>
    <el-row>
        <el-col :span="10" :offset="2">
            <el-form-item prop="enable" :label="$t('website.enable')">
                <el-switch v-model="enableUpdate.enable" @change="updateEnable"></el-switch>
            </el-form-item>
            <el-form-item :label="$t('website.data')">
                <el-input
                    type="textarea"
                    :autosize="{ minRows: 4, maxRows: 8 }"
                    v-model="add"
                    :placeholder="$t('website.wafInputHelper')"
                />
            </el-form-item>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column :label="$t('website.value')" prop="value"></el-table-column>
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
import { GetWafConfig, UpdateWafEnable } from '@/api/modules/website';
import { computed, onMounted, reactive, ref } from 'vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { SaveFileContent } from '@/api/modules/files';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';

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

let loading = ref(false);
let data = ref([]);
let req = ref<Website.WafReq>({
    websiteId: 0,
    key: '',
    rule: 'url',
});
let fileUpdate = reactive({
    path: '',
    content: '',
});
let enableUpdate = ref<Website.WafUpdate>({
    websiteId: 0,
    key: '$UrlDeny',
    enable: false,
});
let add = ref();

const get = async () => {
    data.value = [];
    loading.value = true;
    const res = await GetWafConfig(req.value);
    loading.value = false;
    enableUpdate.value.enable = res.data.enable;
    if (res.data.content != '') {
        const urlList = res.data.content.split('\n');
        urlList.forEach((value) => {
            if (value != '') {
                data.value.push({
                    value: value,
                });
            }
        });
    }
    fileUpdate.path = res.data.filePath;
};

const remove = (index: number) => {
    data.value.splice(index, 1);
    const addArray = [];
    data.value.forEach((d) => {
        addArray.push(d.value);
    });
    submit(addArray);
};

const openCreate = () => {
    const addArray = add.value.split('\n');
    if (addArray.length == 0) {
        return;
    }
    data.value.forEach((d) => {
        addArray.push(d.value);
    });
    submit(addArray);
};

const updateEnable = async (enable: boolean) => {
    enableUpdate.value.enable = enable;
    loading.value = true;
    await UpdateWafEnable(enableUpdate.value);
    loading.value = false;
};

const submit = async (addArray: string[]) => {
    let urlList = '';
    addArray.forEach((row) => {
        urlList = urlList + row + '\n';
    });

    fileUpdate.content = urlList;
    loading.value = true;
    SaveFileContent(fileUpdate)
        .then(() => {
            add.value = '';
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
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
    get();
});
</script>
