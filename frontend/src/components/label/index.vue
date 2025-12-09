<template>
    <div>
        <el-space wrap>
            <div v-for="(env, index) in tmpLabels" :key="index">
                <fu-read-write-switch write-trigger="onDblclick">
                    <template #read>
                        <el-button plain>
                            {{ env }}
                            <el-button
                                circle
                                plain
                                size="small"
                                type="danger"
                                class="float-right -mr-6 -mt-8"
                                icon="Close"
                                @click="remove(index)"
                            />
                        </el-button>
                    </template>
                    <template #default="{ read }">
                        <el-input class="p-w-300" v-model="tmpLabels[index]" @blur="read" />
                    </template>
                </fu-read-write-switch>
            </div>
            <el-input v-if="showAdd" v-model="labelItem">
                <template #append>
                    <el-button icon="Check" @click="save()" />
                    <el-button icon="Close" @click="cancel()" />
                </template>
            </el-input>
            <el-button plain icon="Plus" @click="add()" />
        </el-space>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
const em = defineEmits(['update:labels']);

const tmpLabels = ref([]);
const labelItem = ref('');
const showAdd = ref(false);

const props = defineProps({
    labels: { type: Array<string>, default: [] },
});
watch(
    () => props.labels,
    (newVal) => {
        tmpLabels.value = newVal || [];
    },
);
const add = () => {
    showAdd.value = true;
    labelItem.value = '';
};
const cancel = () => {
    showAdd.value = false;
    labelItem.value = '';
};
const save = () => {
    if (labelItem.value && tmpLabels.value.indexOf(labelItem.value) === -1) {
        tmpLabels.value.push(labelItem.value);
    }
    showAdd.value = false;
    labelItem.value = '';
    handleUpdate();
};
const remove = (index: number) => {
    tmpLabels.value.splice(index, 1);
    handleUpdate();
};
const handleUpdate = () => {
    em('update:labels', tmpLabels.value);
};
onMounted(() => {
    tmpLabels.value = props.labels || [];
});
</script>
