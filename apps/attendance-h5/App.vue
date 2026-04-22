<script>
	import {
		installRouteAuthInterceptors,
		scheduleRouteGuardCheck
	} from "./utils/auth.js";

	export default {
		onLaunch: function() {
			installRouteAuthInterceptors();
			scheduleRouteGuardCheck();
			console.log('App Launch');
		},
		onShow: function() {
			scheduleRouteGuardCheck();
			console.log('App Show');
		},
		onHide: function() {
			console.log('App Hide');
		}
	}
</script>

<style lang="scss">
	/*每个页面公共css */
	@import 'uview-plus/index.scss';
</style>


