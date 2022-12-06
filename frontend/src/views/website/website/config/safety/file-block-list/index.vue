<template>
    <el-row>
        <el-col :span="10" :offset="2">
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </template>
                <el-table-column :label="$t('website.fileExt')">
                    <template #default="{ row }">
                        <fu-read-write-switch :data="row.file" v-model="row.edit" write-trigger="onDblclick">
                            <el-form-item :error="row.error">
                                <el-input v-model="row.file" @blur="row.edit = false" />
                            </el-form-item>
                        </fu-read-write-switch>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')">
                    <template #default="{ $index }">
                        <el-button link type="primary" @click="remove($index)">
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
import { WebSite } from '@/api/interface/website';
import { GetWafConfig } from '@/api/modules/website';
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
});
const id = computed(() => {
    return props.id;
});

let loading = ref(false);
let data = ref([]);
let req = ref<WebSite.WafReq>({
    websiteId: 0,
    key: '',
    rule: 'blackfileExt',
});
let fileUpdate = reactive({
    path: '',
    content: '',
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
                file: value,
                eidt: false,
                error: '',
            });
        });
    }

    fileUpdate.path = res.data.filePath;
};

const remove = (index: number) => {
    data.value.splice(index, 1);
};

const openCreate = () => {
    data.value.unshift({ file: '', edit: true, error: '' });
};

const submit = async () => {
    let fileArray = [];
    data.value.forEach((row) => {
        if (row.file != '') {
            fileArray.push(row.file);
        }
    });

    fileUpdate.content = JSON.stringify(fileArray);
    loading.value = true;
    SaveFileContent(fileUpdate)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
            get();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    req.value.websiteId = id.value;
    get();
});
</script>
