<template>
    <div v-if="persistenceShow">
        <el-card style="margin-top: 5px">
            <el-form :model="form" ref="formRef" :rules="rules" label-width="120px">
                <el-row>
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
                        <el-form>
                            <el-form-item label="appendonly" prop="appendonly">
                                <el-switch v-model="form.appendonly"></el-switch>
                            </el-form-item>
                            <el-form-item label="appendfsync" prop="appendfsync">
                                <el-radio-group v-model="form.appendfsync">
                                    <el-radio label="always">always</el-radio>
                                    <el-radio label="everysec">everysec</el-radio>
                                    <el-radio label="no">no</el-radio>
                                </el-radio-group>
                            </el-form-item>
                        </el-form>
                    </el-col>
                </el-row>
            </el-form>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { RedisPersistenceConf } from '@/api/modules/database';
import { Rules } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';

const form = reactive({
    appendonly: '',
    appendfsync: 'no',
});
const rules = reactive({
    appendonly: [Rules.requiredSelect],
    appendfsync: [Rules.requiredSelect],
});
const formRef = ref<FormInstance>();

const persistenceShow = ref(false);
const acceptParams = (): void => {
    persistenceShow.value = true;
    loadform();
};
const onClose = (): void => {
    persistenceShow.value = false;
};

// const onSave = async (formEl: FormInstance | undefined, key: string) => {
//     if (!formEl) return;
//     const result = await formEl.validateField(key, callback);
//     if (!result) {
//         return;
//     }
//     // let changeForm = {
//     //     paramName: key,
//     //     value: val + '',
//     // };
//     // await updateRedisConf(changeForm);
//     ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
// };
// function callback(error: any) {
//     if (error) {
//         return error.message;
//     } else {
//         return;
//     }
// }

const loadform = async () => {
    const res = await RedisPersistenceConf();
    form.appendonly = res.data?.appendonly;
    form.appendfsync = res.data?.appendfsync;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
