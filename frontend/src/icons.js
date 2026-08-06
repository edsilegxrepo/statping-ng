import { library } from "@fortawesome/fontawesome-svg-core";
import { fab } from "@fortawesome/free-brands-svg-icons";
import { fas } from "@fortawesome/free-solid-svg-icons";
import {
	FontAwesomeIcon,
	FontAwesomeLayers,
	FontAwesomeLayersText,
} from "@fortawesome/vue-fontawesome";

library.add(fas, fab);

export { FontAwesomeIcon, FontAwesomeLayers, FontAwesomeLayersText };
