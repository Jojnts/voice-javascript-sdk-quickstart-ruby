.PHONY: install serve

install:
	bundle install

serve:
	RUBYOPT= bundle exec ruby app.rb -o 0.0.0.0
