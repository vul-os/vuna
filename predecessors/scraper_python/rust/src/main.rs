
extern crate reqwest;
extern crate scraper;

// importation syntax
use scraper::{Html, Selector};

fn main() {
    hn_headlines("https://kadekillary.work/post/webscraping-rust/");
}

fn hn_headlines(url: &str) {

   let mut resp = reqwest::get(url).unwrap();
   assert!(resp.status().is_success());

   let body = resp.text().unwrap();
   // parses string of HTML as a document
   let fragment = Html::parse_document(&body);
   // parses based on a CSS selector
   let category_ = Selector::parse("div[class='product-categories']").unwrap();
   let category = fragment.select(&category_);

   let categories_ = Selector::parse("li").unwrap();
   
   for story in category.select(&categories) {
        // grab the headline text and place into a vector
        let story_txt = story.text().collect::<Vec<_>>();
        println!("{:?}", story_txt);
    }
}