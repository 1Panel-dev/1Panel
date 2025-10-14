<template>
    <div>
        <ComplexTable stripe v-if="tmpVolumes" style="width: 100%" :data="tmpVolumes">
            <el-table-column :label="$t('container.server')" min-width="200">
                <template #default="{ row }">
                    <el-radio-group v-model="row.type">
                        <el-radio-button value="volume">
                            {{ $t('container.volumeOption') }}
                        </el-radio-button>
                        <el-radio-button value="bind">
                            {{ $t('container.hostOption') }}
                        </el-radio-button>
                    </el-radio-group>
                </template>
            </el-table-column>
            <el-table-column :label="$t('container.server') + ' -> ' + $t('container.containerDir')" min-width="400">
                <template #default="{ row }">
                    <el-select v-if="row.type === 'volume'" filterable v-model="row.sourceDir">
                        <div v-for="(item, indexV) of volumeOptions" :key="indexV">
                            <el-tooltip :hide-after="20" :content="item.option" placement="top">
                                <el-option :value="item.option" :label="item.option.substring(0, 30)" />
                            </el-tooltip>
                        </div>
                    </el-select>
                    <el-input v-else v-model="row.sourceDir" />
                    <el-icon><Bottom /></el-icon>
                    <el-input v-model="row.containerDir" />
                </template>
            </el-table-column>
            <el-table-column :label="$t('container.mode')" min-width="180">
                <template #default="{ row }">
                    <el-radio-group v-model="row.mode">
                        <el-radio value="rw">{{ $t('container.modeRW') }}</el-radio>
                        <el-radio value="ro">{{ $t('container.modeR') }}</el-radio>
                    </el-radio-group>
                </template>
            </el-table-column>

            <el-table-column :label="$t('container.sharedLabel')" min-width="180">
                <template #default="{ row }">
                    <el-select v-if="row.type !== 'volume'" popper-class="tall-options" v-model="row.shared">
                        <el-option value="private" :label="$t('container.private')">
                            <div class="title">{{ $t('container.private') }}</div>
                            <div class="description">{{ $t('container.privateHelper') }}</div>
                        </el-option>
                        <el-option value="rprivate" :label="$t('container.rprivate')">
                            <div class="title">{{ $t('container.rprivate') }}</div>
                            <div class="description">{{ $t('container.rprivateHelper') }}</div>
                        </el-option>
                        <el-option value="shared" :label="$t('container.shared')">
                            <div class="title">{{ $t('container.shared') }}</div>
                            <div class="description">{{ $t('container.sharedHelper') }}</div>
                        </el-option>
                        <el-option value="rshared" :label="$t('container.rshared')">
                            <div class="title">{{ $t('container.rshared') }}</div>
                            <div class="description">{{ $t('container.rsharedHelper') }}</div>
                        </el-option>
                        <el-option value="slave" :label="$t('container.slave')">
                            <div class="title">{{ $t('container.slave') }}</div>
                            <div class="description">{{ $t('container.slaveHelper') }}</div>
                        </el-option>
                        <el-option value="rslave" :label="$t('container.rslave')">
                            <div class="title">{{ $t('container.rslave') }}</div>
                            <div class="description">{{ $t('container.rslaveHelper') }}</div>
                        </el-option>
                    </el-select>
                    <span v-else>-</span>
                </template>
            </el-table-column>
            <el-table-column min-width="80" :fixed="'right'">
                <template #default="scope">
                    <el-button link type="primary" @click="handleVolumesDelete(scope.$index)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
            </el-table-column>
        </ComplexTable>
        <el-button class="mt-2" @click="handleVolumesAdd()">
            {{ $t('commons.button.add') }}
        </el-button>
    </div>
</template>

<script lang="ts" setup>
import { Container } from '@/api/interface/container';
import { listVolume } from '@/api/modules/container';
import { ref, onMounted } from 'vue';
const tmpVolumes = ref([]);
const volumeOptions = ref();
const props = defineProps({
    volumes: { type: Array<Container.Volume>, default: [] },
});
watch(
    () => props.volumes,
    (newVal) => {
        tmpVolumes.value = newVal || [];
    },
);

const loadVolumeOptions = async () => {
    const res = await listVolume();
    volumeOptions.value = res.data;
};
const handleVolumesAdd = () => {
    let item = {
        type: 'volume',
        sourceDir: '',
        containerDir: '',
        mode: 'rw',
        shared: 'private',
    };
    tmpVolumes.value.push(item);
};
const handleVolumesDelete = (index: number) => {
    tmpVolumes.value.splice(index, 1);
};
const loadVolumes = () => {
    return tmpVolumes.value;
};

onMounted(() => {
    tmpVolumes.value = props.volumes || [];
    loadVolumeOptions();
});

defineExpose({
    loadVolumes,
});
</script>

<style lang="scss" scoped>
.tall-options {
    .el-select-dropdown__item {
        height: 50px !important;
        padding: 5px 15px !important;
        line-height: 1.5 !important;
        font-size: 14px;

        &.selected {
            font-weight: bold;
        }

        &:hover {
            background-color: #f0f9ff;
        }
    }
    .title {
        font-size: 14px;
        font-weight: 500;
    }
    .description {
        font-size: 12px;
        color: #adb0bc;
    }
}
</style>
