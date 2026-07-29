<template>
    <div class="offset-md-3 offset-lg-4 offset-0 col-lg-4 col-md-6 mt-5">

            <div class="col-12 mb-5 text-center">
        <h1 class="text-uppercase font-weight-bold text-dark mb-0" style="letter-spacing: 6px; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; font-size: 2.2rem;">Platform Monitoring</h1>
        <hr class="mt-2 mb-0" style="width: 100px; border-top: 3px solid #007bff; margin: 0 auto;">
      </div>

      <div class="login_container col-12 p-4">
            <FormLogin/>
        </div>
    </div>
</template>

<script>
const FormLogin = () => import(/* webpackChunkName: "index" */ "@/forms/Login");

export default {
	name: "Login",
	components: {
		FormLogin,
	},
	data() {
		return {};
	},
	mounted() {},
	methods: {},
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
