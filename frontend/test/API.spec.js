import { mount } from "@vue/test-utils";
import API from "@/API";

describe("API Tests", async () => {
	await it("should get core info", async () => {
		const wrapper = mount(API);

		const core = await wrapper.core();
		expect(core).toBe(9);
	});
});
