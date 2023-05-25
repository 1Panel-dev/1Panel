<template>
    <el-row>
        <el-col :xs="24" :sm="18" :md="10" :lg="10" :xl="10">
            <el-form-item prop="enable" :label="$t('website.enable')">
                <el-switch v-model="enableUpdate.enable" @change="updateEnable"></el-switch>
            </el-form-item>
            <el-form-item :label="$t('website.ext')">
                <el-input
                    type="textarea"
                    :autosize="{ minRows: 4, maxRows: 8 }"
                    v-model="exts"
                    :placeholder="$t('website.wafInputHelper')"
                />
            </el-form-item>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column :label="$t('website.fileExt')" prop="file"></el-table-column>
                <el-table-column :label="$t('commons.table.operate')">
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
import { SaveFileContent } from '@/api/modules/files';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});

let loading = ref(false);
let data = ref([]);
let req = ref<Website.WafReq>({
    websiteId: 0,
    key: '$fileExtDeny',
    rule: 'file_ext_block',
});
let fileUpdate = reactive({
    path: '',
    content: '',
});
let enableUpdate = ref<Website.WafUpdate>({
    websiteId: 0,
    key: '$fileExtDeny',
    enable: false,
});
let exts = ref();

const get = async () => {
    data.value = [];
    loading.value = true;
    const res = await GetWafConfig(req.value);
    loading.value = false;

    if (res.data.content != '') {
        const ipList = JSON.parse(res.data.content);
        ipList.forEach((value: string) => {
            data.value.push({
                file: value,
            });
        });
    }

    fileUpdate.path = res.data.filePath;
    enableUpdate.value.enable = res.data.enable;
};

const remove = (index: number) => {
    const copyList = data.value.concat();
    copyList.splice(index, 1);
    const extArray = [];
    copyList.forEach((d) => {
        extArray.push(d.file);
    });
    submit(extArray);
};

const openCreate = () => {
    const extArray = exts.value.split('\n');
    if (extArray.length === 0) {
        return;
    }
    data.value.forEach((d) => {
        extArray.push(d.file);
    });
    submit(extArray);
};

const submit = async (extArray: string[]) => {
    fileUpdate.content = JSON.stringify(extArray);
    loading.value = true;
    SaveFileContent(fileUpdate)
        .then(() => {
            exts.value = '';
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            get();
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
    enableUpdate.value.websiteId = id.value;
    get();
});
</script>
