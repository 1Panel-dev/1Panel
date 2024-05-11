<template>
    <div>
        <el-drawer
            v-model="deleteVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('container.imageDelete')" :back="handleClose" />
            </template>
            <el-form @submit.prevent :model="form" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('container.tag')" prop="tagName">
                            <div style="width: 100%">
                                <el-checkbox
                                    v-model="deleteAll"
                                    :indeterminate="isIndeterminate"
                                    @change="handleCheckAllChange"
                                >
                                    {{ $t('container.removeAll') }}
                                </el-checkbox>
                            </div>
                            <el-checkbox-group v-model="form.deleteTags" @change="handleCheckedChange">
                                <div>
                                    <el-checkbox
                                        style="width: 100%"
                                        v-for="item in form.tags"
                                        :key="item"
                                        :value="item"
                                        :label="item"
                                    />
                                </div>
                            </el-checkbox-group>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="deleteVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" :disabled="form.deleteTags.length === 0" @click="batchDelete()">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

        <OpDialog ref="opRef" @search="onSearch" @cancel="handleClose" />
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { ElForm } from 'element-plus';
import { imageRemove } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import i18n from '@/lang';

const deleteVisible = ref(false);
const form = reactive({
    id: '',
    force: false,
    tags: [] as Array<string>,
    deleteTags: [] as Array<string>,
});

const deleteAll = ref();
const isIndeterminate = ref(true);
const opRef = ref();

interface DialogProps {
    id: string;
    isUsed: boolean;
    tags: Array<string>;
}
const acceptParams = (params: DialogProps) => {
    deleteAll.value = false;
    deleteVisible.value = true;
    form.deleteTags = [];
    form.id = params.id;
    form.tags = params.tags;
    form.force = !params.isUsed;
};
const handleClose = () => {
    deleteVisible.value = false;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const onSearch = () => {
    emit('search');
};

const handleCheckAllChange = (val: boolean) => {
    form.deleteTags = val ? form.tags : [];
    isIndeterminate.value = false;
};
const handleCheckedChange = (value: string[]) => {
    const checkedCount = value.length;
    deleteAll.value = checkedCount === form.tags.length;
    isIndeterminate.value = checkedCount > 0 && checkedCount < form.tags.length;
};

const batchDelete = async () => {
    let names = [];
    let showNames = [];
    if (deleteAll.value) {
        names.push(form.id);
        for (const item of form.deleteTags) {
            showNames.push(item);
        }
    } else {
        for (const item of form.deleteTags) {
            names.push(item);
            showNames.push(item);
        }
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: showNames,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.image'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: imageRemove,
        params: { names: names, force: form.force },
    });
};

defineExpose({
    acceptParams,
});
</script>
