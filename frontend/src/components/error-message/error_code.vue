<template>
    <div class="not-container">
        <img src="@/assets/images/error.svg" class="not-img" :alt="props.code" />
        <div class="not-detail">
            <h2>{{ props.code }}</h2>
            <h4>{{ $t('setting.' + 'error' + props.code) }}</h4>
        </div>
    </div>
</template>

<script setup lang="ts">
const props = defineProps({
    code: String,
});
</script>

<style scoped lang="scss">
.not-container {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    .not-img {
        margin-top: 300px;
    }
    .not-detail {
        margin-top: 300px;
        display: flex;
        flex-direction: column;
        h2,
        h4 {
            padding: 0;
            margin: 0;
        }
        h2 {
            font-size: 60px;
            color: #434e59;
        }
        h4 {
            margin: 30px 0 20px;
            font-size: 19px;
            font-weight: normal;
            color: #848587;
        }
    }
}
</style>
